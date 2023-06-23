package cpu

import (
	"fmt"
	"time"

	psutilcpu "github.com/shirou/gopsutil/v3/cpu"
)

var (
	maxFreq uint64
	quota   float64
)

// Info cpu info.
type Info struct {
	Frequency uint64
	Quota     float64
}

// GetInfo get cpu info.
func GetInfo() Info {
	return Info{
		Frequency: maxFreq,
		Quota:     quota,
	}
}

func GetCpuPercent() float64 {
	// use cpu.Percent to get the total cpu usage percentage
	// pass 0 as the first argument to get a single value
	// pass false as the second argument to get the total percentage
	usage, err := psutilcpu.Percent(500*time.Millisecond, false)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	// cpu.Percent returns a slice of float64, use the first element
	return usage[0]
}

//GetClockTicks get the OS's ticks per second
func GetClockTicks() int {
	// TODO figure out a better alternative for platforms where we're missing cgo
	//
	// TODO Windows. This could be implemented using Win32 QueryPerformanceFrequency().
	// https://msdn.microsoft.com/en-us/library/windows/desktop/ms644905(v=vs.85).aspx
	//
	// An example of its usage can be found here.
	// https://msdn.microsoft.com/en-us/library/windows/desktop/dn553408(v=vs.85).aspx

	return 100
}
