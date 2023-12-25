package datax

import (
	"fmt"
	"testing"
)

func TestDynamicStruct(t *testing.T) {
	pe := NewBuilder().
		AddString("Name", "").
		AddInt64("Age", "").
		Build()
	p := pe.New()
	p.SetString("Name", "你好")
	p.SetInt64("Age", 32)
	fmt.Printf("%+v\n", p)
	fmt.Printf("%T，%+v\n", p.Interface(), p.Interface())
	fmt.Printf("%T，%+v\n", p.Addr(), p.Addr())
}

func TestDynamicStructByHandle(t *testing.T) {
	res, _ := DefaultDynamcHandler.BuildModel([]*IDynamcHandler{
		{Name: "username", Mode: JSON, Type: "string"},
		{Name: "password", Mode: JSON, Type: "string"},
		{Name: "payload", Mode: JSON, Type: "object"},
		{Name: "payload.id", Mode: JSON, Type: "int"},
	})
	res.SetObjectValue("payload.id", int64(1))
	fmt.Println(res.GetObjectValue("payload.id"))
	fmt.Println(res.ToMap(map[string]string{
		"payload":    "object",
		"payload.id": "int",
	}))
}
