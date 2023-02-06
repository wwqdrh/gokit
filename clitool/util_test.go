package clitool

import (
	"fmt"
	"testing"
)

func TestGetValue(t *testing.T) {
	var val = struct {
		Name   string `alias:"name" required:"true"`
		Option bool   `required:"true"`
	}{}

	res, err := GetTagValue(val, "required")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(res)

	res, err = GetTagValue(val, "alias")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(res)
}
