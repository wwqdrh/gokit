package datax

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func IsBlank(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return value.IsNil()
	}
	return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
}

func EmptyValue(mode string) interface{} {
	switch mode {
	case "int":
		return 0
	case "float":
		return float64(0)
	case "bool":
		return false
	case "string":
		return ""
	case "[]int":
		return []int{}
	case "[]float":
		return []float64{}
	case "[]bool":
		return []bool{}
	case "[]string":
		return []string{}
	default:
		return nil
	}
}

// mode maybe string, []string, int, []int, bool, []bool, float, []float
// val maybe a simple str or hava "," contains multi value
// trans the val to correct type
func Str2Value(val string, mode string) (interface{}, error) {
	// 根据mode的不同，选择不同的转换方式
	switch mode {
	case "string":
		// 如果mode是string，直接返回val
		return val, nil
	case "[]string":
		// 如果mode是[]string，使用strings.Split函数将val按逗号分割为一个字符串切片，并返回
		return strings.Split(val, ","), nil
	case "int":
		// 如果mode是int，使用strconv.Atoi函数将val转换为一个整数，并返回
		n, err := strconv.Atoi(val)
		if err != nil {
			// 如果转换失败，返回错误信息
			return nil, err
		}
		return n, nil
	case "[]int":
		// 如果mode是[]int，先使用strings.Split函数将val按逗号分割为一个字符串切片
		strs := strings.Split(val, ",")
		// 创建一个整数切片，长度和容量与字符串切片相同
		nums := make([]int, len(strs))
		for i, str := range strs {
			// 遍历字符串切片，使用strconv.Atoi函数将每个字符串转换为一个整数，并存入整数切片中
			n, err := strconv.Atoi(str)
			if err != nil {
				// 如果转换失败，返回错误信息
				return nil, err
			}
			nums[i] = n
		}
		return nums, nil
	case "bool":
		// 如果mode是bool，使用strconv.ParseBool函数将val转换为一个布尔值，并返回
		b, err := strconv.ParseBool(val)
		if err != nil {
			// 如果转换失败，返回错误信息
			return nil, err
		}
		return b, nil
	case "[]bool":
		// 如果mode是[]bool，先使用strings.Split函数将val按逗号分割为一个字符串切片
		strs := strings.Split(val, ",")
		// 创建一个布尔值切片，长度和容量与字符串切片相同
		bools := make([]bool, len(strs))
		for i, str := range strs {
			// 遍历字符串切片，使用strconv.ParseBool函数将每个字符串转换为一个布尔值，并存入布尔值切片中
			b, err := strconv.ParseBool(str)
			if err != nil {
				return nil, err // 如果转换失败，返回错误信息
			}
			bools[i] = b
		}

		return bools, nil // 返回布尔值切片

	case "float":
		// 如果mode是float，使用strconv.ParseFloat函数将val转换为一个浮点数，并返回

		f, err := strconv.ParseFloat(val, 64) // 使用64位精度

		if err != nil {
			return nil, err // 如果转换失败，返回错误信息

		}

		return f, nil // 返回浮点数

	case "[]float":
		// 如果mode是[]float，先使用strings.Split函数将val按逗号分割为一个字符串切片

		strs := strings.Split(val, ",")

		// 创建一个浮点数切片，长度和容量与字符串切片相同

		floats := make([]float64, len(strs))

		for i, str := range strs {
			// 遍历字符串切片，使用strconv.ParseFloat函数将每个字符串转换为一个浮点数，并存入浮点数切片中

			f, err := strconv.ParseFloat(str, 64) // 使用64位精度

			if err != nil {
				return nil, err // 如果转换失败，返回错误信息

			}

			floats[i] = f
		}

		return floats, nil // 返回浮点数切片

	default:
		// 如果mode不是以上任何一种，返回一个错误信息

		return nil, fmt.Errorf("invalid mode: %s", mode)
	}
}
