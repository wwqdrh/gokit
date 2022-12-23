package datax

import (
	"fmt"
	"testing"
)

func TestGroupJson(t *testing.T) {
	type SimpleGroup struct {
		Name  string `json:"name" groups:"group1"`
		Value string `json:"value" groups:"group2"`
		Both  string `json:"both" groups:"both"`
	}

	v := SimpleGroup{
		Name:  "name1",
		Value: "value1",
		Both:  "both1",
	}

	res, err := Marshal(&Options{
		Groups: []string{"group1"},
	}, v)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)

	res, err = Marshal(&Options{
		Groups: []string{"group1", "group2"},
	}, v)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)

	res, err = Marshal(&Options{}, v)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)

	res, err = Marshal(&Options{Groups: nil}, v)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
}
