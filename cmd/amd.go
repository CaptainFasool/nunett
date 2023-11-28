package cmd

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
)

type amdGPU struct {
	index    int
	executer Executer
}

func (a *amdGPU) Name() string {
	var (
		pattern string
		re      *regexp.Regexp
		match   []string

		cmd Commander
		out []byte

		name string

		err error
	)

	pattern = fmt.Sprintf(`GPU\[%d\]\s+: Card series:\s+(.+)`, a.index)
	re = regexp.MustCompile(pattern)

	cmd = a.executer.Execute("sh", "-c", "rocm-smi --showproductname")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return ""
	}

	match = re.FindStringSubmatch(string(out))
	if len(match) < 1 {
		return ""
	}
	name = match[1]

	return name
}

func (a *amdGPU) UtilizationRate() uint32 {
	var (
		pattern string
		re      *regexp.Regexp
		match   []string

		cmd Commander
		out []byte

		utilization int64

		err error
	)

	pattern = fmt.Sprintf(`GPU\[%d\]\s+: GPU use \(%%\): (\d+)`, a.index)
	re = regexp.MustCompile(pattern)

	cmd = a.executer.Execute("sh", "-c", "rocm-smi --showuse")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return 0
	}

	match = re.FindStringSubmatch(string(out))
	if len(match) < 1 {
		return 0
	}

	utilization, err = strconv.ParseInt(match[1], 10, 32)
	if err != nil {
		return 0
	}
	return uint32(utilization)
}

func (a *amdGPU) Memory() memoryInfo {
	var (
		patternTotal, patternUsed string
		reTotal, reUsed           *regexp.Regexp
		matchTotal, matchUsed     []string

		cmd Commander
		out []byte

		total, used, free int64

		err error
	)

	patternTotal = fmt.Sprintf(`GPU\[%d\]\s+: vis_vram Total Memory \(B\): (\d+)`, a.index)
	reTotal = regexp.MustCompile(patternTotal)

	patternUsed = fmt.Sprintf(`GPU\[%d\]\s+: vis_vram Total Used Memory \(B\): (\d+)`, a.index)
	reUsed = regexp.MustCompile(patternUsed)

	cmd = a.executer.Execute("rocm-smi --showmeminfo vis_vram")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return memoryInfo{}
	}

	matchTotal = reTotal.FindStringSubmatch(string(out))
	matchUsed = reUsed.FindStringSubmatch(string(out))
	if len(matchTotal) < 1 && len(matchUsed) < 1 {
		return memoryInfo{}
	}

	total, err = strconv.ParseInt(matchTotal[1], 10, 64)
	if err != nil {
		total = 0
	}

	used, err = strconv.ParseInt(matchUsed[1], 10, 64)
	if err != nil {
		used = 0
	}

	free = (total - used)
	return memoryInfo{
		used:  uint64(used),
		free:  uint64(free),
		total: uint64(total),
	}
}

func (a *amdGPU) Temperature() float64 {
	var (
		pattern string
		re      *regexp.Regexp
		match   []string

		cmd Commander
		out []byte

		temperature float64

		err error
	)

	pattern = fmt.Sprintf(`GPU\[%d\]\s+: Temperature \(Sensor edge\) \(C\): ([\d\.]+)`, a.index)
	re = regexp.MustCompile(pattern)

	cmd = a.executer.Execute("rocm-smi --showtemp")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return 0
	}

	match = re.FindStringSubmatch(string(out))
	if len(match) < 1 {
		return 0
	}

	temperature, err = strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0
	}

	return temperature
}

func (a *amdGPU) PowerUsage() uint32 {
	var (
		pattern string
		re      *regexp.Regexp
		match   []string

		cmd Commander
		out []byte

		powerFloat float64
		power      uint32

		err error
	)

	pattern = fmt.Sprintf(`GPU\[%d\]\s+: Average Graphics Package Power \(W\): ([\d\.]+)`, a.index)
	re = regexp.MustCompile(pattern)

	cmd = a.executer.Execute("rocm-smi --showpower")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return 0
	}

	match = re.FindStringSubmatch(string(out))
	if len(match) < 1 {
		return 0
	}

	powerFloat, err = strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0
	}
	power = uint32(math.Round(powerFloat))

	return power
}
