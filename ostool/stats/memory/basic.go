package memory

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

func GetBasicMemory() (uint64, uint64) {
	// use mem.VirtualMemory to get the memory information
	memory, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println(err)
		return 0, 0
	}
	return memory.Total, memory.Available
}

func GetTaskMemory(pid int32) (uint64, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return 0, err
	}

	mem, err := p.MemoryInfo()
	return mem.RSS, err
}
