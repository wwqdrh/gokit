package cpu

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
