package memory

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/mem"
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
