package basetool

////////
// 当前的反射
// 1、不能设置方法
// 2、不能有小写字段、embed字段
// 3、struct名字为空
//////

import (
	"errors"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"

	"github.com/wwqdrh/gokit/logger"
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
func (b *Builder) AddField(f reflect.StructField) *Builder {
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
	return &Struct{stu, index}
}

func (b *Builder) AddString(name, tag string) *Builder {
	return b.AddField(reflect.StructField{Name: strings.ToUpper(name), Type: reflect.TypeOf(""), Tag: reflect.StructTag(tag)})
}

func (b *Builder) AddStringArray(name, tag string) *Builder {
	return b.AddField(reflect.StructField{Name: strings.ToUpper(name), Type: reflect.TypeOf([]string{""}), Tag: reflect.StructTag(tag)})
}

func (b *Builder) AddBool(name, tag string) *Builder {
	return b.AddField(reflect.StructField{Name: strings.ToUpper(name), Type: reflect.TypeOf(true), Tag: reflect.StructTag(tag)})
}

func (b *Builder) AddBoolArray(name, tag string) *Builder {
	return b.AddField(reflect.StructField{Name: strings.ToUpper(name), Type: reflect.TypeOf([]bool{true}), Tag: reflect.StructTag(tag)})
}

func (b *Builder) AddInt64(name, tag string) *Builder {
	return b.AddField(reflect.StructField{Name: strings.ToUpper(name), Type: reflect.TypeOf(int64(0)), Tag: reflect.StructTag(tag)})
}

func (b *Builder) AddInt64Array(name, tag string) *Builder {
	return b.AddField(reflect.StructField{Name: strings.ToUpper(name), Type: reflect.TypeOf([]int64{0}), Tag: reflect.StructTag(tag)})
}

func (b *Builder) AddFloat64(name, tag string) *Builder {
	return b.AddField(reflect.StructField{Name: strings.ToUpper(name), Type: reflect.TypeOf(float64(1.2)), Tag: reflect.StructTag(tag)})
}

func (b *Builder) AddFloat64Array(name, tag string) *Builder {
	return b.AddField(reflect.StructField{Name: strings.ToUpper(name), Type: reflect.TypeOf([]float64{1.2}), Tag: reflect.StructTag(tag)})
}

func (b *Builder) AddStruct(name string, v interface{}, tag string, annomus bool) *Builder {
	return b.AddField(reflect.StructField{Name: strings.ToUpper(name), Type: reflect.TypeOf(v), Tag: reflect.StructTag(tag), Anonymous: annomus})
}

// 实际生成的结构体，基类
// 结构体的类型
type Struct struct {
	typ reflect.Type
	// <fieldName : 索引> // 用于通过字段名称，从Builder的[]reflect.StructField中获取reflect.StructField
	index map[string]int
}

func (s Struct) New() *Instance {
	return &Instance{reflect.New(s.typ).Elem(), s.index, map[string]bool{}}
}

// 结构体的值
type Instance struct {
	instance reflect.Value
	// <fieldName : 索引>
	index      map[string]int
	valueExist map[string]bool
}

func (in *Instance) ToMap(dataType map[string]string) map[string]interface{} {
	res := map[string]interface{}{}

	for field, fieldType := range dataType {
		if strings.Contains(field, ".") {
			result, err := in.GetValue(field)
			if err != nil {
				logger.DefaultLogger.Warn(err.Error())
				continue
			}
			res[field] = result
		}

		val, err := in.Field(field)
		if err != nil {
			logger.DefaultLogger.Warn(err.Error())
			continue
		}

		switch fieldType {
		case "string":
			res[field] = val.String()
		case "bool":
			res[field] = val.Bool()
		case "int":
			res[field] = val.Int()
		case "float":
			res[field] = val.Float()
		case "[]int":
			res[field] = val.Interface().([]int64)
		case "[]bool":
			res[field] = val.Interface().([]bool)
		case "[]float":
			res[field] = val.Interface().([]float64)
		case "*multipart.FileHeader":
			res[field] = val.Interface().(*multipart.FileHeader)
		case "[]*multipart.FileHeader":
			res[field] = val.Interface().([]*multipart.FileHeader)
		case "object", "[]object":
			result, err := in.GetValue(field)
			if err != nil {
				logger.DefaultLogger.Warn(err.Error())
				continue
			}
			res[field] = result
		default:
			logger.DefaultLogger.Warn("dont support this type")
		}
	}

	return res
}

func (in Instance) Field(name string) (reflect.Value, error) {
	if i, ok := in.index[strings.ToUpper(name)]; ok {
		return in.instance.Field(i), nil
	} else {
		return reflect.Value{}, ErrFieldNoExist
	}
}

// 添加一个方法，不知道什么类型就直接用这个
func (in *Instance) SetValue(name string, value interface{}) {
	in.valueExist[name] = true
	if strings.Contains(name, ".") {
		in.setObjectValue(name, value)
	} else {
		if i, ok := in.index[strings.ToUpper(name)]; ok {
			in.instance.Field(i).Set(reflect.ValueOf(value))
		}
	}
}

func (in *Instance) HasValue(name string) bool {
	_, ok := in.valueExist[name]
	return ok
}

// payload.id
func (in *Instance) setObjectValue(name string, value interface{}) error {
	parts := strings.Split(strings.ToUpper(name), ".")
	var rootField reflect.Value
	if i, ok := in.index[strings.ToUpper(parts[0])]; ok {
		rootField = in.instance.Field(i)
	} else {
		return errors.New("not found this object")
	}
	in.setFieldValue(rootField, parts[1:], value)
	return nil
}

func (in *Instance) setFieldValue(field reflect.Value, parts []string, value interface{}) {
	if len(parts) == 0 {
		field.Set(reflect.ValueOf(value))
		return
	}

	curname := parts[0]
	if field.Kind() == reflect.Array || field.Kind() == reflect.Slice {
		curidx, err := strconv.ParseInt(curname[1:len(curname)-1], 10, 64)
		if err != nil {
			return
		}
		curlen := curidx + 1
		if field.Len() < int(curlen) {
			// 新增
			for i := field.Len(); i < int(curlen)-1; i++ {
				field.Set(reflect.Append(field, reflect.Zero(field.Type().Elem())))
			}
			curitem := reflect.Zero(field.Type().Elem())
			if len(parts) == 1 {
				if value == nil {
					// use default value
					field.Set(reflect.Append(field, curitem))
				} else {
					field.Set(reflect.Append(field, reflect.ValueOf(value)))
				}
			} else {
				field.Set(reflect.Append(field, curitem))
				in.setFieldValue(curitem, parts[1:], value)
			}
		} else {
			// 修改
			idx := int(curidx)
			if idx < field.Len() {
				curitem := field.Index(idx)
				if len(parts) == 1 {
					curitem.Set(reflect.ValueOf(value))
				} else {
					in.setFieldValue(curitem, parts[1:], value)
				}
			}
		}
	} else {
		nextField := field.FieldByName(curname)
		if nextField.IsValid() {
			in.setFieldValue(nextField, parts[1:], value)
		}
	}
}

func (in *Instance) GetValue(name string) (interface{}, error) {
	parts := strings.Split(strings.ToUpper(name), ".")
	var rootField reflect.Value
	if i, ok := in.index[strings.ToUpper(parts[0])]; ok {
		rootField = in.instance.Field(i)
	} else {
		return nil, errors.New("not found this object")
	}

	curfield, err := in.getFieldValue(rootField, parts[1:])
	if err != nil {
		return nil, err
	}
	// 如果curfield是一个struct的话，递归将该struct转为map[string]interface{}
	if curfield.Kind() == reflect.Struct {
		return in.structToMap(curfield), nil
	} else if curfield.Kind() == reflect.Array || curfield.Kind() == reflect.Slice {
		if curfield.Len() > 0 && curfield.Index(0).Kind() == reflect.Struct {
			res := []map[string]interface{}{}
			for i := 0; i < curfield.Len(); i++ {
				res = append(res, in.structToMap(curfield.Index(i)))
			}
			return res, nil
		}
	}

	return curfield.Interface(), nil
}

func (in *Instance) getFieldValue(field reflect.Value, parts []string) (reflect.Value, error) {
	if !field.IsValid() {
		return reflect.Value{}, errors.New("invalid field")
	}

	if len(parts) == 0 {
		return field, nil
	}

	nextField := field.FieldByName(parts[0])
	if !nextField.IsValid() {
		return reflect.Value{}, errors.New("invalid field")
	}

	return in.getFieldValue(nextField, parts[1:])
}

func (in *Instance) structToMap(structField reflect.Value) map[string]interface{} {
	dataMap := make(map[string]interface{})
	for i := 0; i < structField.NumField(); i++ {
		key := structField.Type().Field(i).Name
		val := structField.Field(i).Interface()

		if structField.Field(i).Kind() == reflect.Struct {
			val = in.structToMap(structField.Field(i))
		}

		dataMap[key] = val
	}

	return dataMap
}

func (i *Instance) Interface() interface{} {
	return i.instance.Interface()
}

func (i *Instance) Type() reflect.Type {
	return i.instance.Type()
}

func (i *Instance) Addr() interface{} {
	return i.instance.Addr().Interface()
}
