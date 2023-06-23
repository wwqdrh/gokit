package storage

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/disk"
)

func GetBasicStorage() map[string][2]uint64 {
	// use disk.Partitions to get the disk partitions information
	partitions, err := disk.Partitions(false)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	result := make(map[string][2]uint64)
	for _, partition := range partitions {
		// use disk.Usage to get the disk usage information for each partition
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			fmt.Println(err)
			continue
		}
		result[partition.Device] = [2]uint64{usage.Total, usage.Free}
	}
	return result
}
