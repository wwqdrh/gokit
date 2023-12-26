package basetool

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Error values.
var (
	ErrNoField         = errors.New("specified field is not present in the struct")
	ErrNotPtr          = errors.New("specified struct is not passed by pointer")
	ErrNotStruct       = errors.New("given object is not a struct or a pointer to a struct")
	ErrUnexportedField = errors.New("specified field is not an exported or public field")
	ErrMismatchValue   = errors.New("specified value to set is of a different type")
)

// GetValue returns the value of a given field of a structure given by 'obj'.
// 'obj' can be passed by value or by pointer.
// Only exported (public) field values can be found (else ErrUnexportedField is raised).
//
// If the field is not found, then an error is returned.
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

func GetFieldIter(obj interface{}, prefix string) ([]string, error) {
	fields, err := Kinds(obj)
	if err != nil {
		return nil, err
	}

	res := []string{}
	for item := range fields {
		if strings.Index(item, prefix) == 0 {
			res = append(res, item)
		}
	}
	return res, nil
}

func GetMethod(obj interface{}, methodName string) (interface{}, error) {
	objValue, err := getReflectValue(obj)
	if err != nil {
		return nil, err
	}

	fieldValue := objValue.MethodByName(methodName)
	if !fieldValue.IsValid() {
		return nil, ErrNoField
	}

	if !fieldValue.CanInterface() {
		return nil, ErrUnexportedField
	}

	return fieldValue.Interface(), nil
}

// Has returns a boolean indicating if the given field name is found in
// the given struct obj.
func Has(obj interface{}, fieldName string) (bool, error) {
	objValue, err := getReflectValue(obj)
	if err != nil {
		return false, err
	}

	structType := objValue.Type()
	_, found := structType.FieldByName(fieldName)
	return found, nil
}

// SetValue sets the given value to the fieldName field in the given struct 'obj'.
// Only exported (public) fields can be set using this API.
//
// NOTE: 'obj' struct must be passed by pointer for this API to work. Passing by
// value results in ErrPassedByValue.
func SetValue(obj interface{}, fieldName string, newValue interface{}) error {
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() != reflect.Ptr {
		return ErrNotPtr
	}

	objValue = objValue.Elem()
	if objValue.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	fieldValue := objValue.FieldByName(fieldName)
	if !fieldValue.IsValid() {
		return ErrNoField
	}

	if fieldValue.Type() != reflect.TypeOf(newValue) {
		return ErrMismatchValue
	}

	if !fieldValue.CanSet() {
		return ErrUnexportedField
	}

	fieldValue.Set(reflect.ValueOf(newValue))
	return nil
}

// Names returns a slice of all field names of a given struct.
// Only the exportable (public) field names are returned.
func Names(obj interface{}) ([]string, error) {
	objValue, err := getReflectValue(obj)
	if err != nil {
		return nil, err
	}

	fieldNames := []string{}
	objType := objValue.Type()
	for i := 0; i < objValue.NumField(); i++ {
		fieldType := objType.Field(i)
		fieldValue := objValue.Field(i)

		if fieldValue.CanInterface() {
			fieldNames = append(fieldNames, fieldType.Name)
		}
	}

	return fieldNames, nil
}

// Values returns a map of all field names with the value of each field.
// Only the exportable (public) field name-value pairs are returned.
func Values(obj interface{}) (map[string]interface{}, error) {
	objValue, err := getReflectValue(obj)
	if err != nil {
		return nil, err
	}

	valueMap := map[string]interface{}{}
	objType := objValue.Type()
	for i := 0; i < objValue.NumField(); i++ {
		fieldType := objType.Field(i)
		fieldValue := objValue.Field(i)

		if fieldValue.CanInterface() {
			valueMap[fieldType.Name] = fieldValue.Interface()
		}
	}

	return valueMap, nil
}

// GetTag returns the value of a specified tag on a specified struct field.
// Specified field must be an exportable (public) filed of the struct.
func GetTag(obj interface{}, fieldName, tagKey string) (string, error) {
	objValue, err := getReflectValue(obj)
	if err != nil {
		return "", err
	}

	structType := objValue.Type()
	field, found := structType.FieldByName(fieldName)
	if !found {
		return "", ErrNoField
	}

	if field.PkgPath != "" {
		return "", ErrUnexportedField
	}

	return field.Tag.Get(tagKey), nil
}

// GetKind returns the "kind" of a specified public struct field. "Kind" is
// the in-built type of a variable, such as Uint64, Slice, Struct, Ptr, etc.
func GetKind(obj interface{}, fieldName string) (string, error) {
	objValue, err := getReflectValue(obj)
	if err != nil {
		return "", err
	}

	fieldValue := objValue.FieldByName(fieldName)
	if !fieldValue.IsValid() {
		return "", ErrNoField
	}

	if !fieldValue.CanInterface() {
		return "", ErrUnexportedField
	}

	return fieldValue.Kind().String(), nil
}

// Kinds returns the 'kind' of all the public fields of a struct. "Kind" is
// the in-built type of a variable, such as Uint64, Slice, Struct, Ptr, etc.
func Kinds(obj interface{}) (map[string]string, error) {
	objValue, err := getReflectValue(obj)
	if err != nil {
		return nil, err
	}

	kindMap := map[string]string{}
	objType := objValue.Type()
	for i := 0; i < objValue.NumField(); i++ {
		fieldType := objType.Field(i)
		fieldValue := objValue.Field(i)

		if fieldValue.CanInterface() {
			kindMap[fieldType.Name] = fieldValue.Kind().String()
		}
	}

	return kindMap, nil
}

// getReflectValue gets a reflect-value of a given struct. If it is a pointer
// to a struct, then it gives the reflect-value of the underlying structure.
//
// Returns an error if the given obj is not a struct or a pointer to a struct.
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

// ErrNotAStructPointer indicates that we were expecting a pointer to a struct,
// but got something else.
type ErrNotAStructPointer string

func newErrNotAStructPointer(v interface{}) ErrNotAStructPointer {
	return ErrNotAStructPointer(fmt.Sprintf("%t", v))
}

// Error implements the error interface.
func (e ErrNotAStructPointer) Error() string {
	return fmt.Sprintf("expected a struct, instead got a %T", string(e))
}

// ErrorUnsettable is used when a field cannot be set.
type ErrorUnsettable string

// Error implements the error interface.
func (e ErrorUnsettable) Error() string {
	return fmt.Sprintf("can't set field %s", string(e))
}

// ErrorUnsupportedType indicates that the type of the struct field is not yet
// support in this package.
type ErrorUnsupportedType struct {
	t reflect.Type
}

// Error implements the error interface.
func (e ErrorUnsupportedType) Error() string {
	return fmt.Sprintf("unsupported type %v", e.t)
}

// Apply parses a struct pointer for `default` tags. If the default tag is
// set and the struct member has a default value, the default value will be
// set no the member. Parse expects a struct pointer.
func LoadDefault(t interface{}) error {
	// Make sure we've been given a pointer.
	val := reflect.ValueOf(t)
	if val.Kind() != reflect.Ptr {
		return newErrNotAStructPointer(t)
	}

	// Make sure the pointer is pointing to a struct.
	ref := val.Elem()
	if ref.Kind() != reflect.Struct {
		return newErrNotAStructPointer(t)
	}

	return parseFields(ref)
}

func parseFields(v reflect.Value) error {
	for i := 0; i < v.NumField(); i++ {
		err := parseField(v.Field(i), v.Type().Field(i))
		if err != nil {
			return err
		}
	}
	return nil
}

func parseField(value reflect.Value, field reflect.StructField) error {
	tagVal := field.Tag.Get("default")

	isStruct := value.Kind() == reflect.Struct
	isStructPointer := value.Kind() == reflect.Ptr && value.Type().Elem().Kind() == reflect.Struct

	if (tagVal == "" || tagVal == "-") && !(isStruct || isStructPointer) {
		return nil
	}

	// 不是空就不需要设置， 但是对于struct可能存在部分是空的，需要额外判断
	if !value.IsZero() && value.Kind() != reflect.Struct {
		return nil
	}

	if !value.CanSet() {
		return ErrorUnsettable(field.Name)
	}

	switch value.Kind() {
	case reflect.String:
		value.SetString(tagVal)
		return nil

	case reflect.Bool:
		b, err := strconv.ParseBool(tagVal)
		if err != nil {
			return err
		}
		value.SetBool(b)
		return nil

	case reflect.Int:
		i, err := strconv.ParseInt(tagVal, 10, 32)
		if err != nil {
			return err
		}
		value.SetInt(i)
		return nil

	case reflect.Int8:
		i, err := strconv.ParseInt(tagVal, 10, 8)
		if err != nil {
			return err
		}
		value.SetInt(i)
		return nil

	case reflect.Int16:
		i, err := strconv.ParseInt(tagVal, 10, 16)
		if err != nil {
			return err
		}
		value.SetInt(i)
		return nil

	// NB: int32 is also an alias for a rune
	case reflect.Int32:
		i, err := parseInt32(tagVal)
		if err != nil {
			return err
		}
		value.SetInt(int64(i))
		return nil

	case reflect.Int64:
		i, err := strconv.ParseInt(tagVal, 10, 64)
		if err != nil {
			return err
		}
		value.SetInt(i)
		return nil

	case reflect.Uint:
		i, err := strconv.ParseInt(tagVal, 10, 32)
		if err != nil {
			return err
		}
		value.SetUint(uint64(i))
		return nil

	case reflect.Uint8:
		i, err := strconv.ParseInt(tagVal, 10, 8)
		if err != nil {
			return err
		}
		value.SetUint(uint64(i))
		return nil

	case reflect.Uint16:
		i, err := strconv.ParseInt(tagVal, 10, 16)
		if err != nil {
			return err
		}
		value.SetUint(uint64(i))
		return nil

	case reflect.Uint32:
		i, err := strconv.ParseInt(tagVal, 10, 32)
		if err != nil {
			return err
		}
		value.SetUint(uint64(i))
		return nil

	case reflect.Uint64:
		i, err := strconv.ParseInt(tagVal, 10, 64)
		if err != nil {
			return err
		}
		value.SetUint(uint64(i))
		return nil

	case reflect.Float32:
		f, err := strconv.ParseFloat(tagVal, 32)
		if err != nil {
			return err
		}
		value.SetFloat(f)
		return nil

	case reflect.Float64:
		f, err := strconv.ParseFloat(tagVal, 64)
		if err != nil {
			return err
		}
		value.SetFloat(f)
		return nil

	case reflect.Slice:
		switch value.Type().Elem().Kind() {
		// a []uint8 is a an alias for a []byte
		case reflect.Uint8:
			value.SetBytes([]byte(tagVal))
			return nil
		case reflect.Int64:
			res := []int64{}
			for _, item := range strings.Split(tagVal, ",") {
				v, err := strconv.ParseInt(item, 10, 64)
				if err != nil {
					continue
				}
				res = append(res, v)
			}
			val := reflect.MakeSlice(reflect.TypeOf([]int64{}), len(res), len(res))
			for i, v := range res {
				val.Index(i).Set(reflect.ValueOf(v))
			}
			value.Set(val)
			return nil

		case reflect.Int:
			res := []int{}
			for _, item := range strings.Split(tagVal, ",") {
				v, err := strconv.ParseInt(item, 10, 64)
				if err != nil {
					continue
				}
				res = append(res, int(v))
			}
			val := reflect.MakeSlice(reflect.TypeOf([]int{}), len(res), len(res))
			for i, v := range res {
				val.Index(i).Set(reflect.ValueOf(v))
			}
			value.Set(val)
			return nil

		case reflect.String:
			res := strings.Split(tagVal, ",")
			val := reflect.MakeSlice(reflect.TypeOf([]string{}), len(res), len(res))
			for i, v := range res {
				val.Index(i).Set(reflect.ValueOf(v))
			}
			value.Set(val)
			return nil

		default:
			return ErrorUnsupportedType{value.Type()}
		}

	case reflect.Struct:
		if value.NumField() == 0 {
			return nil
		}
		return parseFields(value) // recurse

	case reflect.Ptr:
		ref := value.Type().Elem()

		switch ref.Kind() {
		case reflect.String:
			value.Set(reflect.ValueOf(&tagVal))
			return nil

		case reflect.Bool:
			b, err := strconv.ParseBool(tagVal)
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(&b))
			return nil

		case reflect.Int:
			n, err := strconv.ParseInt(tagVal, 10, 32)
			if err != nil {
				return err
			}
			i := int(n)
			value.Set(reflect.ValueOf(&i))
			return nil

		case reflect.Int8:
			n, err := strconv.ParseInt(tagVal, 10, 8)
			if err != nil {
				return err
			}
			i := int8(n)
			value.Set(reflect.ValueOf(&i))
			return nil

		case reflect.Int16:
			n, err := strconv.ParseInt(tagVal, 10, 16)
			if err != nil {
				return err
			}
			i := int16(n)
			value.Set(reflect.ValueOf(&i))
			return nil

		case reflect.Int32:
			// NB: *int32 is an alias for a *rune
			i, err := parseInt32(tagVal)
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(&i))
			return nil

		case reflect.Int64:
			i, err := strconv.ParseInt(tagVal, 10, 64)
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(&i))
			return nil

		case reflect.Uint:
			n, err := strconv.ParseInt(tagVal, 10, 32)
			if err != nil {
				return err
			}
			u := uint(n)
			value.Set(reflect.ValueOf(&u))
			return nil

		case reflect.Uint8:
			n, err := strconv.ParseInt(tagVal, 10, 8)
			if err != nil {
				return err
			}
			u := uint8(n)
			value.Set(reflect.ValueOf(&u))
			return nil

		case reflect.Uint16:
			n, err := strconv.ParseInt(tagVal, 10, 16)
			if err != nil {
				return err
			}
			u := uint16(n)
			value.Set(reflect.ValueOf(&u))
			return nil

		case reflect.Uint32:
			n, err := strconv.ParseInt(tagVal, 10, 32)
			if err != nil {
				return err
			}
			u := uint32(n)
			value.Set(reflect.ValueOf(&u))
			return nil

		case reflect.Uint64:
			n, err := strconv.ParseInt(tagVal, 10, 64)
			if err != nil {
				return err
			}
			u := uint64(n)
			value.Set(reflect.ValueOf(&u))
			return nil

		case reflect.Float32:
			f, err := strconv.ParseFloat(tagVal, 32)
			if err != nil {
				return err
			}
			f32 := float32(f)
			value.Set(reflect.ValueOf(&f32))
			return nil

		case reflect.Float64:
			f, err := strconv.ParseFloat(tagVal, 64)
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(&f))
			return nil

		case reflect.Slice:
			switch ref.Elem().Kind() {
			// a *[]uint is an alias for *[]byte
			case reflect.Uint8:
				b := []byte(tagVal)
				value.Set(reflect.ValueOf(&b))
				return nil

			default:
				return ErrorUnsupportedType{value.Type()}
			}

		case reflect.Struct:
			if ref.NumField() == 0 {
				return nil
			}

			// If it's nil set it to it's default value so we can set the
			// children if we need to.
			if value.IsNil() {
				value.Set(reflect.New(ref))
			}
			return parseFields(value.Elem()) // recurse

		default:
			return ErrorUnsupportedType{value.Type()}
		}

	default:
		return ErrorUnsupportedType{value.Type()}
	}
}

// Attempt to parse a string as an int32 and, failing that, a rune.
func parseInt32(s string) (int32, error) {
	// Try parsing it as an int.
	i, err := strconv.ParseInt(s, 10, 32)
	if err == nil {
		return int32(i), nil
	}

	// We couldn't parse it as an int, maybe it's a rune.
	runes := []rune(s)
	if len(runes) == 1 {
		return runes[0], nil
	} else {
		return 0, err
	}
}
