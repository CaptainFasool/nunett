function nextTapDevice() {
    counter=-1
    while [ $? -eq 0 ]; do
        counter=$(($counter+1))
        ip link show tap$counter &> /dev/null 
    done

    echo $counter
}

# nextTapDevice

function nextIPRange() {
    for ((i=0; i<$(nextTapDevice);++i)); do
        output=$(ip -br addr show tap$i | awk '{print $3}')
    done
    echo $output | awk -F. '{ print $1"."$2"."$3+1"."$4 }'
}

nextIPRange
