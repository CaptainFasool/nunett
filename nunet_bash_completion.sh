# bash completion for nunet                                -*- shell-script -*-

__nunet_debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE:-} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    fi
}

# Homebrew on Macs have version 1.3 of bash-completion which doesn't include
# _init_completion. This is a very minimal version of that function.
__nunet_init_completion()
{
    COMPREPLY=()
    _get_comp_words_by_ref "$@" cur prev words cword
}

__nunet_index_of_word()
{
    local w word=$1
    shift
    index=0
    for w in "$@"; do
        [[ $w = "$word" ]] && return
        index=$((index+1))
    done
    index=-1
}

__nunet_contains_word()
{
    local w word=$1; shift
    for w in "$@"; do
        [[ $w = "$word" ]] && return
    done
    return 1
}

__nunet_handle_go_custom_completion()
{
    __nunet_debug "${FUNCNAME[0]}: cur is ${cur}, words[*] is ${words[*]}, #words[@] is ${#words[@]}"

    local shellCompDirectiveError=1
    local shellCompDirectiveNoSpace=2
    local shellCompDirectiveNoFileComp=4
    local shellCompDirectiveFilterFileExt=8
    local shellCompDirectiveFilterDirs=16

    local out requestComp lastParam lastChar comp directive args

    # Prepare the command to request completions for the program.
    # Calling ${words[0]} instead of directly nunet allows to handle aliases
    args=("${words[@]:1}")
    # Disable ActiveHelp which is not supported for bash completion v1
    requestComp="NUNET_ACTIVE_HELP=0 ${words[0]} __completeNoDesc ${args[*]}"

    lastParam=${words[$((${#words[@]}-1))]}
    lastChar=${lastParam:$((${#lastParam}-1)):1}
    __nunet_debug "${FUNCNAME[0]}: lastParam ${lastParam}, lastChar ${lastChar}"

    if [ -z "${cur}" ] && [ "${lastChar}" != "=" ]; then
        # If the last parameter is complete (there is a space following it)
        # We add an extra empty parameter so we can indicate this to the go method.
        __nunet_debug "${FUNCNAME[0]}: Adding extra empty parameter"
        requestComp="${requestComp} \"\""
    fi

    __nunet_debug "${FUNCNAME[0]}: calling ${requestComp}"
    # Use eval to handle any environment variables and such
    out=$(eval "${requestComp}" 2>/dev/null)

    # Extract the directive integer at the very end of the output following a colon (:)
    directive=${out##*:}
    # Remove the directive
    out=${out%:*}
    if [ "${directive}" = "${out}" ]; then
        # There is not directive specified
        directive=0
    fi
    __nunet_debug "${FUNCNAME[0]}: the completion directive is: ${directive}"
    __nunet_debug "${FUNCNAME[0]}: the completions are: ${out}"

    if [ $((directive & shellCompDirectiveError)) -ne 0 ]; then
        # Error code.  No completion.
        __nunet_debug "${FUNCNAME[0]}: received error from custom completion go code"
        return
    else
        if [ $((directive & shellCompDirectiveNoSpace)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __nunet_debug "${FUNCNAME[0]}: activating no space"
                compopt -o nospace
            fi
        fi
        if [ $((directive & shellCompDirectiveNoFileComp)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __nunet_debug "${FUNCNAME[0]}: activating no file completion"
                compopt +o default
            fi
        fi
    fi

    if [ $((directive & shellCompDirectiveFilterFileExt)) -ne 0 ]; then
        # File extension filtering
        local fullFilter filter filteringCmd
        # Do not use quotes around the $out variable or else newline
        # characters will be kept.
        for filter in ${out}; do
            fullFilter+="$filter|"
        done

        filteringCmd="_filedir $fullFilter"
        __nunet_debug "File filtering command: $filteringCmd"
        $filteringCmd
    elif [ $((directive & shellCompDirectiveFilterDirs)) -ne 0 ]; then
        # File completion for directories only
        local subdir
        # Use printf to strip any trailing newline
        subdir=$(printf "%s" "${out}")
        if [ -n "$subdir" ]; then
            __nunet_debug "Listing directories in $subdir"
            __nunet_handle_subdirs_in_dir_flag "$subdir"
        else
            __nunet_debug "Listing directories in ."
            _filedir -d
        fi
    else
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${out}" -- "$cur")
    fi
}

__nunet_handle_reply()
{
    __nunet_debug "${FUNCNAME[0]}"
    local comp
    case $cur in
        -*)
            if [[ $(type -t compopt) = "builtin" ]]; then
                compopt -o nospace
            fi
            local allflags
            if [ ${#must_have_one_flag[@]} -ne 0 ]; then
                allflags=("${must_have_one_flag[@]}")
            else
                allflags=("${flags[*]} ${two_word_flags[*]}")
            fi
            while IFS='' read -r comp; do
                COMPREPLY+=("$comp")
            done < <(compgen -W "${allflags[*]}" -- "$cur")
            if [[ $(type -t compopt) = "builtin" ]]; then
                [[ "${COMPREPLY[0]}" == *= ]] || compopt +o nospace
            fi

            # complete after --flag=abc
            if [[ $cur == *=* ]]; then
                if [[ $(type -t compopt) = "builtin" ]]; then
                    compopt +o nospace
                fi

                local index flag
                flag="${cur%=*}"
                __nunet_index_of_word "${flag}" "${flags_with_completion[@]}"
                COMPREPLY=()
                if [[ ${index} -ge 0 ]]; then
                    PREFIX=""
                    cur="${cur#*=}"
                    ${flags_completion[${index}]}
                    if [ -n "${ZSH_VERSION:-}" ]; then
                        # zsh completion needs --flag= prefix
                        eval "COMPREPLY=( \"\${COMPREPLY[@]/#/${flag}=}\" )"
                    fi
                fi
            fi

            if [[ -z "${flag_parsing_disabled}" ]]; then
                # If flag parsing is enabled, we have completed the flags and can return.
                # If flag parsing is disabled, we may not know all (or any) of the flags, so we fallthrough
                # to possibly call handle_go_custom_completion.
                return 0;
            fi
            ;;
    esac

    # check if we are handling a flag with special work handling
    local index
    __nunet_index_of_word "${prev}" "${flags_with_completion[@]}"
    if [[ ${index} -ge 0 ]]; then
        ${flags_completion[${index}]}
        return
    fi

    # we are parsing a flag and don't have a special handler, no completion
    if [[ ${cur} != "${words[cword]}" ]]; then
        return
    fi

    local completions
    completions=("${commands[@]}")
    if [[ ${#must_have_one_noun[@]} -ne 0 ]]; then
        completions+=("${must_have_one_noun[@]}")
    elif [[ -n "${has_completion_function}" ]]; then
        # if a go completion function is provided, defer to that function
        __nunet_handle_go_custom_completion
    fi
    if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
        completions+=("${must_have_one_flag[@]}")
    fi
    while IFS='' read -r comp; do
        COMPREPLY+=("$comp")
    done < <(compgen -W "${completions[*]}" -- "$cur")

    if [[ ${#COMPREPLY[@]} -eq 0 && ${#noun_aliases[@]} -gt 0 && ${#must_have_one_noun[@]} -ne 0 ]]; then
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${noun_aliases[*]}" -- "$cur")
    fi

    if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
        if declare -F __nunet_custom_func >/dev/null; then
            # try command name qualified custom func
            __nunet_custom_func
        else
            # otherwise fall back to unqualified for compatibility
            declare -F __custom_func >/dev/null && __custom_func
        fi
    fi

    # available in bash-completion >= 2, not always present on macOS
    if declare -F __ltrim_colon_completions >/dev/null; then
        __ltrim_colon_completions "$cur"
    fi

    # If there is only 1 completion and it is a flag with an = it will be completed
    # but we don't want a space after the =
    if [[ "${#COMPREPLY[@]}" -eq "1" ]] && [[ $(type -t compopt) = "builtin" ]] && [[ "${COMPREPLY[0]}" == --*= ]]; then
       compopt -o nospace
    fi
}

# The arguments should be in the form "ext1|ext2|extn"
__nunet_handle_filename_extension_flag()
{
    local ext="$1"
    _filedir "@(${ext})"
}

__nunet_handle_subdirs_in_dir_flag()
{
    local dir="$1"
    pushd "${dir}" >/dev/null 2>&1 && _filedir -d && popd >/dev/null 2>&1 || return
}

__nunet_handle_flag()
{
    __nunet_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    # if a command required a flag, and we found it, unset must_have_one_flag()
    local flagname=${words[c]}
    local flagvalue=""
    # if the word contained an =
    if [[ ${words[c]} == *"="* ]]; then
        flagvalue=${flagname#*=} # take in as flagvalue after the =
        flagname=${flagname%=*} # strip everything after the =
        flagname="${flagname}=" # but put the = back
    fi
    __nunet_debug "${FUNCNAME[0]}: looking for ${flagname}"
    if __nunet_contains_word "${flagname}" "${must_have_one_flag[@]}"; then
        must_have_one_flag=()
    fi

    # if you set a flag which only applies to this command, don't show subcommands
    if __nunet_contains_word "${flagname}" "${local_nonpersistent_flags[@]}"; then
      commands=()
    fi

    # keep flag value with flagname as flaghash
    # flaghash variable is an associative array which is only supported in bash > 3.
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        if [ -n "${flagvalue}" ] ; then
            flaghash[${flagname}]=${flagvalue}
        elif [ -n "${words[ $((c+1)) ]}" ] ; then
            flaghash[${flagname}]=${words[ $((c+1)) ]}
        else
            flaghash[${flagname}]="true" # pad "true" for bool flag
        fi
    fi

    # skip the argument to a two word flag
    if [[ ${words[c]} != *"="* ]] && __nunet_contains_word "${words[c]}" "${two_word_flags[@]}"; then
        __nunet_debug "${FUNCNAME[0]}: found a flag ${words[c]}, skip the next argument"
        c=$((c+1))
        # if we are looking for a flags value, don't show commands
        if [[ $c -eq $cword ]]; then
            commands=()
        fi
    fi

    c=$((c+1))

}

__nunet_handle_noun()
{
    __nunet_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    if __nunet_contains_word "${words[c]}" "${must_have_one_noun[@]}"; then
        must_have_one_noun=()
    elif __nunet_contains_word "${words[c]}" "${noun_aliases[@]}"; then
        must_have_one_noun=()
    fi

    nouns+=("${words[c]}")
    c=$((c+1))
}

__nunet_handle_command()
{
    __nunet_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    local next_command
    if [[ -n ${last_command} ]]; then
        next_command="_${last_command}_${words[c]//:/__}"
    else
        if [[ $c -eq 0 ]]; then
            next_command="_nunet_root_command"
        else
            next_command="_${words[c]//:/__}"
        fi
    fi
    c=$((c+1))
    __nunet_debug "${FUNCNAME[0]}: looking for ${next_command}"
    declare -F "$next_command" >/dev/null && $next_command
}

__nunet_handle_word()
{
    if [[ $c -ge $cword ]]; then
        __nunet_handle_reply
        return
    fi
    __nunet_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"
    if [[ "${words[c]}" == -* ]]; then
        __nunet_handle_flag
    elif __nunet_contains_word "${words[c]}" "${commands[@]}"; then
        __nunet_handle_command
    elif [[ $c -eq 0 ]]; then
        __nunet_handle_command
    elif __nunet_contains_word "${words[c]}" "${command_aliases[@]}"; then
        # aliashash variable is an associative array which is only supported in bash > 3.
        if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
            words[c]=${aliashash[${words[c]}]}
            __nunet_handle_command
        else
            __nunet_handle_noun
        fi
    else
        __nunet_handle_noun
    fi
    __nunet_handle_word
}

_nunet_autocomplete()
{
    last_command="nunet_autocomplete"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")
    local_nonpersistent_flags+=("--help")
    local_nonpersistent_flags+=("-h")

    must_have_one_flag=()
    must_have_one_noun=()
    must_have_one_noun+=("bash")
    must_have_one_noun+=("zsh")
    noun_aliases=()
}

_nunet_capacity()
{
    last_command="nunet_capacity"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--available")
    flags+=("-a")
    local_nonpersistent_flags+=("--available")
    local_nonpersistent_flags+=("-a")
    flags+=("--full")
    flags+=("-f")
    local_nonpersistent_flags+=("--full")
    local_nonpersistent_flags+=("-f")
    flags+=("--onboarded")
    flags+=("-o")
    local_nonpersistent_flags+=("--onboarded")
    local_nonpersistent_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_chat_clear()
{
    last_command="nunet_chat_clear"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_chat_join()
{
    last_command="nunet_chat_join"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_chat_list()
{
    last_command="nunet_chat_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_chat_start()
{
    last_command="nunet_chat_start"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_chat()
{
    last_command="nunet_chat"

    command_aliases=()

    commands=()
    commands+=("clear")
    commands+=("join")
    commands+=("list")
    commands+=("start")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_device_set()
{
    last_command="nunet_device_set"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_device_status()
{
    last_command="nunet_device_status"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_device()
{
    last_command="nunet_device"

    command_aliases=()

    commands=()
    commands+=("set")
    commands+=("status")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_gpu_capacity()
{
    last_command="nunet_gpu_capacity"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--cuda-tensor")
    flags+=("-c")
    local_nonpersistent_flags+=("--cuda-tensor")
    local_nonpersistent_flags+=("-c")
    flags+=("--rocm-hip")
    flags+=("-r")
    local_nonpersistent_flags+=("--rocm-hip")
    local_nonpersistent_flags+=("-r")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_gpu_onboard()
{
    last_command="nunet_gpu_onboard"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_gpu_status()
{
    last_command="nunet_gpu_status"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_gpu()
{
    last_command="nunet_gpu"

    command_aliases=()

    commands=()
    commands+=("capacity")
    commands+=("onboard")
    commands+=("status")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_help()
{
    last_command="nunet_help"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    has_completion_function=1
    noun_aliases=()
}

_nunet_info()
{
    last_command="nunet_info"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_log()
{
    last_command="nunet_log"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_offboard()
{
    last_command="nunet_offboard"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_onboard()
{
    last_command="nunet_onboard"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--address=")
    two_word_flags+=("--address")
    two_word_flags+=("-a")
    local_nonpersistent_flags+=("--address")
    local_nonpersistent_flags+=("--address=")
    local_nonpersistent_flags+=("-a")
    flags+=("--cardano")
    flags+=("-C")
    local_nonpersistent_flags+=("--cardano")
    local_nonpersistent_flags+=("-C")
    flags+=("--cpu=")
    two_word_flags+=("--cpu")
    two_word_flags+=("-c")
    local_nonpersistent_flags+=("--cpu")
    local_nonpersistent_flags+=("--cpu=")
    local_nonpersistent_flags+=("-c")
    flags+=("--local-enable")
    flags+=("-l")
    local_nonpersistent_flags+=("--local-enable")
    local_nonpersistent_flags+=("-l")
    flags+=("--memory=")
    two_word_flags+=("--memory")
    two_word_flags+=("-m")
    local_nonpersistent_flags+=("--memory")
    local_nonpersistent_flags+=("--memory=")
    local_nonpersistent_flags+=("-m")
    flags+=("--ntx-price=")
    two_word_flags+=("--ntx-price")
    two_word_flags+=("-x")
    local_nonpersistent_flags+=("--ntx-price")
    local_nonpersistent_flags+=("--ntx-price=")
    local_nonpersistent_flags+=("-x")
    flags+=("--nunet-channel=")
    two_word_flags+=("--nunet-channel")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--nunet-channel")
    local_nonpersistent_flags+=("--nunet-channel=")
    local_nonpersistent_flags+=("-n")
    flags+=("--plugin=")
    two_word_flags+=("--plugin")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--plugin")
    local_nonpersistent_flags+=("--plugin=")
    local_nonpersistent_flags+=("-p")
    flags+=("--unavailable")
    flags+=("-u")
    local_nonpersistent_flags+=("--unavailable")
    local_nonpersistent_flags+=("-u")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_onboard-ml()
{
    last_command="nunet_onboard-ml"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_peer_list()
{
    last_command="nunet_peer_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--dht")
    flags+=("-d")
    local_nonpersistent_flags+=("--dht")
    local_nonpersistent_flags+=("-d")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_peer_self()
{
    last_command="nunet_peer_self"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_peer()
{
    last_command="nunet_peer"

    command_aliases=()

    commands=()
    commands+=("list")
    commands+=("self")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_resource-config()
{
    last_command="nunet_resource-config"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--cpu=")
    two_word_flags+=("--cpu")
    two_word_flags+=("-c")
    local_nonpersistent_flags+=("--cpu")
    local_nonpersistent_flags+=("--cpu=")
    local_nonpersistent_flags+=("-c")
    flags+=("--memory=")
    two_word_flags+=("--memory")
    two_word_flags+=("-m")
    local_nonpersistent_flags+=("--memory")
    local_nonpersistent_flags+=("--memory=")
    local_nonpersistent_flags+=("-m")
    flags+=("--ntx-price=")
    two_word_flags+=("--ntx-price")
    two_word_flags+=("-x")
    local_nonpersistent_flags+=("--ntx-price")
    local_nonpersistent_flags+=("--ntx-price=")
    local_nonpersistent_flags+=("-x")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_run()
{
    last_command="nunet_run"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_shell()
{
    last_command="nunet_shell"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--node-id=")
    two_word_flags+=("--node-id")
    local_nonpersistent_flags+=("--node-id")
    local_nonpersistent_flags+=("--node-id=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_version()
{
    last_command="nunet_version"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_wallet_new()
{
    last_command="nunet_wallet_new"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--cardano")
    flags+=("-c")
    local_nonpersistent_flags+=("--cardano")
    local_nonpersistent_flags+=("-c")
    flags+=("--ethereum")
    flags+=("-e")
    local_nonpersistent_flags+=("--ethereum")
    local_nonpersistent_flags+=("-e")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_wallet()
{
    last_command="nunet_wallet"

    command_aliases=()

    commands=()
    commands+=("new")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_nunet_root_command()
{
    last_command="nunet"

    command_aliases=()

    commands=()
    commands+=("autocomplete")
    commands+=("capacity")
    commands+=("chat")
    commands+=("device")
    commands+=("gpu")
    commands+=("help")
    commands+=("info")
    commands+=("log")
    commands+=("offboard")
    commands+=("onboard")
    commands+=("onboard-ml")
    commands+=("peer")
    commands+=("resource-config")
    commands+=("run")
    commands+=("shell")
    commands+=("version")
    commands+=("wallet")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

__start_nunet()
{
    local cur prev words cword split
    declare -A flaghash 2>/dev/null || :
    declare -A aliashash 2>/dev/null || :
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -s || return
    else
        __nunet_init_completion -n "=" || return
    fi

    local c=0
    local flag_parsing_disabled=
    local flags=()
    local two_word_flags=()
    local local_nonpersistent_flags=()
    local flags_with_completion=()
    local flags_completion=()
    local commands=("nunet")
    local command_aliases=()
    local must_have_one_flag=()
    local must_have_one_noun=()
    local has_completion_function=""
    local last_command=""
    local nouns=()
    local noun_aliases=()

    __nunet_handle_word
}

if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_nunet nunet
else
    complete -o default -o nospace -F __start_nunet nunet
fi

# ex: ts=4 sw=4 et filetype=sh
