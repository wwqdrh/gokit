package clitool

import (
	"errors"
	"reflect"
)

var (
	ErrNoField         = errors.New("specified field is not present in the struct")
	ErrNotPtr          = errors.New("specified struct is not passed by pointer")
	ErrNotStruct       = errors.New("given object is not a struct or a pointer to a struct")
	ErrUnexportedField = errors.New("specified field is not an exported or public field")
	ErrMismatchValue   = errors.New("specified value to set is of a different type")
)

func GetValue(obj interface{}, fieldName string) (interface{}, error) {
	objValue, err := getReflectValue(obj)
	if err != nil {
		return nil, err
	}

	fieldValue := objValue.FieldByName(fieldName)
	if !fieldValue.IsValid() {
		return nil, ErrNoField
	}

	if !fieldValue.CanInterface() {
		return nil, ErrUnexportedField
	}

	return fieldValue.Interface(), nil
}

func GetTagValue(obj interface{}, tagKey string) (map[string]string, error) {
	objValue, err := getReflectValue(obj)
	if err != nil {
		return nil, err
	}

	tagMap := map[string]string{}
	objType := objValue.Type()
	for i := 0; i < objValue.NumField(); i++ {
		fieldType := objType.Field(i)
		fieldValue := objValue.Field(i)

		if fieldValue.CanInterface() {
			tagMap[fieldType.Name] = fieldType.Tag.Get(tagKey)
		}
	}

	return tagMap, nil
}

func getReflectValue(obj interface{}) (reflect.Value, error) {
	value := reflect.ValueOf(obj)

	if value.Kind() == reflect.Struct {
		return value, nil
	}

	if value.Kind() == reflect.Ptr && value.Elem().Kind() == reflect.Struct {
		return value.Elem(), nil
	}

	var retval reflect.Value
	return retval, ErrNotStruct
}
