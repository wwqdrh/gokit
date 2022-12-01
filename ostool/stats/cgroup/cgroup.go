package cgroup

import (
	"fmt"
	"os"
	"path"
	"strings"
)

const cgroupRootDir = "/sys/fs/cgroup"

type cgroup struct {
	cgroupSet map[string]string
}

// example: proc/[pid]/cgroup: 0::/system.slice/docker-2c461f8a586f1903cba222b65dfdcaadf60094b29b344539d7b9fbd9d3abe2c2.scope
func NewCgroup(pid int) (*cgroup, error) {
	cgroupFile := fmt.Sprintf("/proc/%d/cgroup", pid)
	data, err := os.ReadFile(cgroupFile)
	if err != nil {
		return nil, err
	}
	dir := strings.Split(string(data), "::")[1]

	f, err := os.Open(path.Join(cgroupRootDir, dir))
	if err != nil {
		return nil, err
	}
	files, err := f.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	set := map[string]string{}
	for _, item := range files {
		set[item.Name()] = path.Join(f.Name(), item.Name())
	}
	return &cgroup{
		cgroupSet: set,
	}, nil
}
