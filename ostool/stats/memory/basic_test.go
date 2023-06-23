package memory

import (
	"fmt"
	"testing"
)

func TestGetBasicMemory(t *testing.T) {
	total, usage := GetBasicMemory()
	fmt.Println(total, usage)
}
