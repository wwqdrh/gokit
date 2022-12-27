package datax

////////
// 当前的反射
// 1、不能设置方法
// 2、不能有小写字段、embed字段
// 3、struct名字为空
//////

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrFieldNoExist = errors.New("field no exist")
)

// 构造器
type Builder struct {
	// 用于存储属性字段
	fileId []reflect.StructField
}

func NewBuilder() *Builder {
	return &Builder{}
}

// 添加字段
func (b *Builder) AddField(field string, f reflect.StructField) *Builder {
	b.fileId = append(b.fileId, f)
	return b
}

// 根据预先添加的字段构建出结构体
func (b *Builder) Build() *Struct {
	stu := reflect.StructOf(b.fileId)

	index := make(map[string]int)
	for i := 0; i < stu.NumField(); i++ {
		index[stu.Field(i).Name] = i
	}
	fmt.Println(stu.Name())
	return &Struct{stu, index}
}

func (b *Builder) AddString(name, tag string) *Builder {
	return b.AddField(name, reflect.StructField{Name: name, Type: reflect.TypeOf(""), Tag: reflect.StructTag(tag)})
}

func (b *Builder) AddBool(name, tag string) *Builder {
	return b.AddField(name, reflect.StructField{Name: name, Type: reflect.TypeOf(true), Tag: reflect.StructTag(tag)})
}

func (b *Builder) AddInt64(name, tag string) *Builder {
	return b.AddField(name, reflect.StructField{Name: name, Type: reflect.TypeOf(int64(0)), Tag: reflect.StructTag(tag)})
}

func (b *Builder) AddFloat64(name, tag string) *Builder {
	return b.AddField(name, reflect.StructField{Name: name, Type: reflect.TypeOf(float64(1.2)), Tag: reflect.StructTag(tag)})
}

func (b *Builder) AddStruct(name string, v interface{}, tag string, annomus bool) *Builder {
	return b.AddField(name, reflect.StructField{Name: name, Type: reflect.TypeOf(v), Tag: reflect.StructTag(tag), Anonymous: annomus})
}

// 实际生成的结构体，基类
// 结构体的类型
type Struct struct {
	typ reflect.Type
	// <fieldName : 索引> // 用于通过字段名称，从Builder的[]reflect.StructField中获取reflect.StructField
	index map[string]int
}

func (s Struct) New() *Instance {
	return &Instance{reflect.New(s.typ).Elem(), s.index}
}

// 结构体的值
type Instance struct {
	instance reflect.Value
	// <fieldName : 索引>
	index map[string]int
}

func (in Instance) Field(name string) (reflect.Value, error) {
	if i, ok := in.index[name]; ok {
		return in.instance.Field(i), nil
	} else {
		return reflect.Value{}, ErrFieldNoExist
	}
}

func (in *Instance) SetString(name, value string) {
	if i, ok := in.index[name]; ok {
		in.instance.Field(i).SetString(value)
	}
}

func (in *Instance) SetBool(name string, value bool) {
	if i, ok := in.index[name]; ok {
		in.instance.Field(i).SetBool(value)
	}
}

func (in *Instance) SetInt64(name string, value int64) {
	if i, ok := in.index[name]; ok {
		in.instance.Field(i).SetInt(value)
	}
}

func (in *Instance) SetFloat64(name string, value float64) {
	if i, ok := in.index[name]; ok {
		in.instance.Field(i).SetFloat(value)
	}
}

func (i *Instance) Interface() interface{} {
	return i.instance.Interface()
}

func (i *Instance) Addr() interface{} {
	return i.instance.Addr().Interface()
}
