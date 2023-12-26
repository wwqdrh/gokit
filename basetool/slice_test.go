package basetool

import (
	"fmt"
	"testing"
	"time"
)

func TestSafeStringSlice(t *testing.T) {
	slice := (SafeSlice[string]{}).NewSafeSlice(10)

	go func() {
		slice.Add("1")
	}()
	go func() {
		slice.Add("2")
	}()
	go func() {
		slice.Add("3")
	}()
	go func() {
		data, err := slice.Get()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(data)
	}()
	go func() {
		data, err := slice.Get()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(data)
	}()
	time.Sleep(3 * time.Second)
}
