// for http
package basetool

import (
	"encoding/json"
	"fmt"
	"math"
	"mime/multipart"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/wwqdrh/gokit/logger"
)

type ReqType uint8

const (
	Nil ReqType = iota
	JSON
	XML
	Form
	Query
	FormPost
	FormMultipart
	ProtoBuf
	MsgPack
	YAML
	Uri
	Header
	TOML
)

const (
	MIMEJSON              = "application/json"
	MIMEHTML              = "text/html"
	MIMEXML               = "application/xml"
	MIMEXML2              = "text/xml"
	MIMEPlain             = "text/plain"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm = "multipart/form-data"
	MIMEPROTOBUF          = "application/x-protobuf"
	MIMEMSGPACK           = "application/x-msgpack"
	MIMEMSGPACK2          = "application/msgpack"
	MIMEYAML              = "application/x-yaml"
	MIMETOML              = "application/toml"
)

func NewReqType(mode string) ReqType {
	switch mode {
	case "query":
		return Query
	case "form":
		return Form
	case "json":
		return JSON
	case "header":
		return Header
	case "uri":
		return Uri
	default:
		return Query
	}
}

func (t ReqType) String() string {
	switch t {
	case Query:
		return "query"
	case Form:
		return "form"
	case XML:
		return "xml"
	case JSON:
		return "json"
	case Uri:
		return "uri"
	default:
		return ""
	}
}

var DefaultDynamcHandler = IDynamcHandler{}

type IDynamcHandler struct {
	Name     string
	Type     string
	Mode     ReqType
	Required bool
	Default  interface{}
	Validate string
	visited  bool

	children []*IDynamcHandler // only for json mode
}

// FromJsonStr from json str to
func (r IDynamcHandler) FromJsonData(data map[string]interface{}, validate map[string]interface{}) []*IDynamcHandler {
	res := []*IDynamcHandler{}
	for key, val := range data {
		curvalidatestr := ""
		nextvalidate := map[string]interface{}{}
		if validate[key] != nil {
			switch v := validate[key].(type) {
			case string:
				curvalidatestr = v
			case map[string]interface{}:
				nextvalidate = v
				if v := v["_"]; v != nil {
					if v, ok := v.(string); ok {
						curvalidatestr = v
					}
				}
			}
		}

		switch v := val.(type) {
		case string:
			res = append(res, &IDynamcHandler{
				Name:     key,
				Type:     "string",
				Mode:     JSON,
				Validate: curvalidatestr,
			})
		case int64, int32, int, uint64, uint32, uint:
			res = append(res, &IDynamcHandler{
				Name:     key,
				Type:     "int",
				Mode:     JSON,
				Validate: curvalidatestr,
			})
		case float64, float32:
			res = append(res, &IDynamcHandler{
				Name:     key,
				Type:     "float",
				Mode:     JSON,
				Validate: curvalidatestr,
			})
		case bool:
			res = append(res, &IDynamcHandler{
				Name:     key,
				Type:     "bool",
				Mode:     JSON,
				Validate: curvalidatestr,
			})
		case map[string]interface{}:
			res = append(res, &IDynamcHandler{
				Name:     key,
				Type:     "-",
				Mode:     JSON,
				Validate: curvalidatestr,
				children: r.FromJsonData(v, nextvalidate),
			})
		case []map[string]interface{}:
			if len(v) == 0 {
				continue
			}
			res = append(res, &IDynamcHandler{
				Name:     key,
				Type:     "array",
				Mode:     JSON,
				Validate: curvalidatestr,
				children: r.FromJsonData(v[0], nextvalidate),
			})
		case []interface{}:
			if len(v) == 0 {
				continue
			}
			if vMap, ok := v[0].(map[string]interface{}); ok {
				res = append(res, &IDynamcHandler{
					Name:     key,
					Type:     "array",
					Mode:     JSON,
					Validate: curvalidatestr,
					children: r.FromJsonData(vMap, nextvalidate),
				})
			}
		}
	}
	return res
}

func (r IDynamcHandler) BuildModel(prefix string, request []*IDynamcHandler) (*Instance, string) {
	sort.Slice(request, func(i, j int) bool {
		return request[i].Name < request[j].Name
	})

	var contentType string

	mod := NewBuilder()
	for _, item := range request {
		if item.visited {
			continue
		}
		item.visited = true

		var tag string

		itemName := strings.TrimLeft(item.Name, prefix)
		if itemName == "" {
			continue
		}
		if itemName[0] == '.' {
			itemName = itemName[1:]
		}

		validatestr := item.Validate
		nextvalidate := map[string]interface{}{}
		if err := json.Unmarshal([]byte(item.Validate), &nextvalidate); err != nil {
			logger.DefaultLogger.Debug(err.Error())
		}

		switch item.Mode {
		case Query:
			tag = fmt.Sprintf(`%s:"%s" validate:"%s"`, "query", itemName, item.Validate)
		case Form:
			tag = fmt.Sprintf(`%s:"%s" validate:"%s"`, "form", itemName, item.Validate)
			if contentType == "" {
				contentType = MIMEMultipartPOSTForm
			}
		case Header:
			tag = fmt.Sprintf(`%s:"%s" validate:"%s"`, "header", itemName, item.Validate)
		case Uri:
			tag = fmt.Sprintf(`%s:"%s" validate:"%s"`, "uri", itemName, item.Validate)
		case JSON:
			if len(nextvalidate) == 0 {
				// plain text
				tag = fmt.Sprintf(`%s:"%s" validate:"%s"`, "form", itemName, validatestr)
			} else {
				// json text
				if v := nextvalidate["_"]; v != nil {
					if v, ok := v.(string); ok {
						tag = fmt.Sprintf(`%s:"%s" validate:"%s"`, "form", itemName, v)
					}
				}
			}
			if contentType != "" {
				contentType = MIMEJSON
			}
		default:
			tag = fmt.Sprintf(`%s:"%s" validate:"%s"`, "form", itemName, item.Validate)
		}
		if item.Required {
			tag += " required: true"
		}
		if item.Default != nil {
			tag += fmt.Sprintf(" default: %v", item.Default)
		}

		switch item.Type {
		case "datetime":
			mod = mod.AddDatetime(itemName, tag)
		case "date":
			mod = mod.AddDate(itemName, tag)
		case "string":
			mod = mod.AddString(itemName, tag)
		case "[]string":
			mod = mod.AddStringArray(itemName, tag)
		case "int":
			mod = mod.AddInt64(itemName, tag)
		case "[]int":
			mod = mod.AddInt64Array(itemName, tag)
		case "float":
			mod = mod.AddFloat64(itemName, tag)
		case "[]float":
			mod = mod.AddFloat64Array(itemName, tag)
		case "bool":
			mod = mod.AddBool(itemName, tag)
		case "[]bool":
			mod = mod.AddBoolArray(itemName, tag)
		case "file":
			f := &multipart.FileHeader{}
			mod = mod.AddStruct(itemName, f, tag, false)
		case "object":
			m, _ := r.BuildModelByPrefix(item.Name, request)
			mod = mod.AddStruct(itemName, reflect.New(m.Type()).Elem().Interface(), tag, false)
		case "[]object":
			m, _ := r.BuildModelByPrefix(item.Name, request)
			mod = mod.AddStruct(itemName, reflect.Zero(reflect.SliceOf(m.Type())).Interface(), tag, false)
		default:
			// maybe a json mod
			if len(item.children) > 0 {
				if item.Type == "array" {
					m, _ := r.BuildModel("", item.children)
					mod = mod.AddStruct(itemName, reflect.Zero(reflect.SliceOf(m.Type())).Interface(), tag, false)
				} else {
					m, _ := r.BuildModel("", item.children)
					mod = mod.AddStruct(itemName, reflect.New(m.Type()).Elem().Interface(), tag, false)
				}
				continue
			}

			var jsondata map[string]interface{}
			if err := json.Unmarshal([]byte(item.Type), &jsondata); err == nil {
				m, _ := r.BuildModel("", r.FromJsonData(jsondata, nextvalidate))
				mod = mod.AddStruct(itemName, reflect.New(m.Type()).Elem().Interface(), tag, false)
				continue
			}

			// todo the []map[string]interface{} type
			var jsondataArr []map[string]interface{}
			if err := json.Unmarshal([]byte(item.Type), &jsondataArr); err == nil && len(jsondataArr) == 1 {
				m, _ := r.BuildModel("", r.FromJsonData(jsondataArr[0], nextvalidate))
				mod = mod.AddStruct(itemName, reflect.Zero(reflect.SliceOf(m.Type())).Interface(), tag, false)
				continue
			}

			logger.DefaultLogger.Warn("不支持该数据类型")
		}
	}
	ins := mod.Build().New()
	for _, item := range request {
		if item.Default != nil {
			ins.SetValue(item.Name, item.Default)
		}
	}
	return ins, contentType
}

func (r IDynamcHandler) BuildModelByPrefix(prefix string, request []*IDynamcHandler) (*Instance, string) {
	handles := []*IDynamcHandler{}
	for _, item := range request {
		if item.visited {
			continue
		}
		if strings.HasPrefix(item.Name, prefix) {
			curname := strings.TrimLeft(item.Name, prefix)
			if strings.HasPrefix(curname, ".") {
				handles = append(handles, item)
			}
		}
	}
	return r.BuildModel(prefix, handles)
}

func (r IDynamcHandler) BindValue(request []*IDynamcHandler, getVal func(item *IDynamcHandler) (interface{}, error)) (*Instance, error) {
	res, _ := r.BuildModel("", request)
	var errs error
	for _, item := range request {
		val, err := getVal(item)
		if err != nil {
			if errs == nil {
				errs = err
			} else {
				errs = errors.Wrapf(err, "%s\n", err.Error())
			}
			continue
		}

		switch item.Type {
		case "string":
			res.SetValue(item.Name, fmt.Sprint(val))
		case "[]string":
			if cv, ok := val.([]string); ok {
				res.SetValue(item.Name, cv)
				continue
			} else if cv, ok := val.([]interface{}); ok {
				curs := []string{}
				for _, item := range cv {
					curs = append(curs, fmt.Sprint(item))
				}
				res.SetValue(item.Name, curs)
				continue
			}
			logger.DefaultLogger.Warn("not a []string type")
		case "int":
			if cv, err := strconv.ParseInt(fmt.Sprint(val), 10, 64); err != nil {
				logger.DefaultLogger.Warn("not a int64")
			} else {
				res.SetValue(item.Name, cv)
			}
		case "[]int":
			if cv, ok := val.([]int64); ok {
				res.SetValue(item.Name, cv)
				continue
			} else if cv, ok := val.([]int32); ok {
				res.SetValue(item.Name, cv)
				continue
			} else if cv, ok := val.([]int); ok {
				res.SetValue(item.Name, cv)
				continue
			} else if cv, ok := val.([]interface{}); ok {
				curs := []int{}
				for _, item := range cv {
					if vint, ok := item.(int); ok {
						curs = append(curs, vint)
					} else if vint, ok := item.(int64); ok {
						curs = append(curs, int(vint))
					}
				}
				res.SetValue(item.Name, curs)
				continue
			}
			logger.DefaultLogger.Warn("not a []string type")
		case "float":
			if cv, ok := val.(float64); !ok {
				logger.DefaultLogger.Warn("not a float")
				res.SetValue(item.Name, float64(0))
			} else {
				res.SetValue(item.Name, cv)
			}
		case "[]float":
			if cv, ok := val.([]float64); ok {
				res.SetValue(item.Name, cv)
				continue
			} else if cv, ok := val.([]interface{}); ok {
				curs := []float64{}
				for _, item := range cv {
					if vint, ok := item.(float64); ok {
						curs = append(curs, vint)
					}
				}
				res.SetValue(item.Name, curs)
				continue
			}
			logger.DefaultLogger.Warn("not a []float type")
		case "bool":
			if cv, ok := val.(bool); ok {
				res.SetValue(item.Name, cv)
				continue
			}
			if cv, ok := val.(string); ok {
				if cv == "true" {
					res.SetValue(item.Name, true)
					continue
				} else if cv == "false" {
					res.SetValue(item.Name, false)
					continue
				}
			}
			logger.DefaultLogger.Warn("not a bool")
		case "[]bool":
			if cv, ok := val.([]bool); ok {
				res.SetValue(item.Name, cv)
				continue
			} else if cv, ok := val.([]interface{}); ok {
				curs := []bool{}
				for _, item := range cv {
					if vint, ok := item.(bool); ok {
						curs = append(curs, vint)
					}
				}
				res.SetValue(item.Name, curs)
				continue
			}
			logger.DefaultLogger.Warn("not a []bool type")
		case "file":
			if cv, ok := val.(*multipart.FileHeader); ok {
				res.SetValue(item.Name, cv)
				continue
			}
			logger.DefaultLogger.Warn("not a file type")
		case "[]file":
			if cv, ok := val.([]*multipart.FileHeader); ok {
				res.SetValue(item.Name, cv)
				continue
			}
			logger.DefaultLogger.Warn("not a []file type")
		case "datetime", "date":
			// 默认是以秒为单位
			switch cv := val.(type) {
			case int64:
				res.SetValue(item.Name, time.Unix(int64(cv), 0))
			case int:
				res.SetValue(item.Name, time.Unix(int64(cv), 0))
			case float64:
				res.SetValue(item.Name, time.Unix(int64(cv), 0))
			case string:
				// 如果是纯数字字符串
				if val, err := strconv.ParseInt(cv, 10, 64); err == nil {
					if len(cv) > 10 {
						val /= int64(math.Pow10(len(cv) - 10))
					} else if len(cv) < 10 {
						val *= int64(math.Pow10(10 - len(cv)))
					}
					res.SetValue(item.Name, time.Unix(int64(val), 0))
				} else {
					t, err := time.Parse(time.RFC3339, cv)
					if err != nil {
						logger.DefaultLogger.Warn(err.Error())
					} else {
						res.SetValue(item.Name, t)
					}
				}
			}
		}
	}

	return res, nil
}
