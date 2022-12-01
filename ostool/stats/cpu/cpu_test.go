package cpu

import (
	"fmt"
	"testing"
)

func TestUsageSysetmCPU(t *testing.T) {
	if usage, err := StatSystemCPUUsage(); err != nil {
		t.Error(err.Error())
	} else {
		fmt.Println(usage)
	}
}

func TestStatCpuFreq(t *testing.T) {
	fmt.Println(StatCpuFreq())
}

func TestStatCpuMaxFreq(t *testing.T) {
	fmt.Println(StatCpuMaxFreq())
}
