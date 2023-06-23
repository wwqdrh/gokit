package storage

import (
	"fmt"
	"testing"
)

func TestBasicStorage(t *testing.T) {
	res := GetBasicStorage()
	fmt.Println(res)
}
