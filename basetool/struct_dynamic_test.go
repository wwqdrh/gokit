package basetool

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestDynamicStruct(t *testing.T) {
	pe := NewBuilder().
		AddString("Name", "").
		AddInt64("Age", "").
		Build()
	p := pe.New()
	p.SetValue("Name", "你好")
	p.SetValue("Age", int64(32))
	fmt.Printf("%+v\n", p)
	fmt.Printf("%T，%+v\n", p.Interface(), p.Interface())
	fmt.Printf("%T，%+v\n", p.Addr(), p.Addr())
}

func TestDynamicStructByHandle(t *testing.T) {
	requests := []*IDynamcHandler{
		{Name: "username", Mode: JSON, Type: "string"},
		{Name: "password", Mode: JSON, Type: "string"},
		{Name: "payload", Mode: JSON, Type: "object"},
		{Name: "payload.id", Mode: JSON, Type: "int"},
		{Name: "extra", Mode: JSON, Type: "[]object"},
		{Name: "extra.id", Mode: JSON, Type: "int"},
		{Name: "ids", Mode: JSON, Type: "[]int"},
	}
	res, _ := DefaultDynamcHandler.BuildModel("", requests)

	res.SetValue("payload.id", int64(1))
	fmt.Println(res.GetValue("payload.id"))

	res.SetValue("ids.[0]", int64(1))
	res.SetValue("ids.[1]", int64(2))
	res.SetValue("ids.[0]", int64(3))
	fmt.Println(res.GetValue("ids"))

	res.SetValue("extra.[0]", nil) // 必须先设置
	res.SetValue("extra.[1]", nil)
	res.SetValue("extra.[0].id", int64(1))
	res.SetValue("extra.[1].id", int64(2))

	fmt.Println(res.ToMap(map[string]string{
		"payload":    "object",
		"payload.id": "int",
		"ids":        "[]int",
		"extra":      "[]object",
	}))
}

func TestDyncmicHandlerBindValue(t *testing.T) {
	h := []*IDynamcHandler{
		{
			Name: "create_at", Mode: JSON, Type: "datetime",
		},
	}

	handler := &IDynamcHandler{}
	res, err := handler.BindValue(h, func(item *IDynamcHandler) (interface{}, error) {
		if item.Name == "create_at" {
			return 1737820800, nil
		}
		return nil, errors.New("not val")
	})
	if err != nil {
		t.Error(err)
		return
	}
	create_at, err := res.GetValue("create_at")
	if err != nil {
		t.Error(err)
		return
	}
	create_at_val, ok := create_at.(time.Time)
	if !ok {
		t.Error("create_at not a time.Time")
		return
	}
	if tstr := fmt.Sprintf("%d-%d-%d", create_at_val.Year(), create_at_val.Month(), create_at_val.Day()); tstr != "2025-1-26" {
		t.Error("时间错误, " + tstr)
		return
	}
}
