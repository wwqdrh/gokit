// for http
package datax

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"reflect"
	"strconv"

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
	Validate string

	children []IDynamcHandler // only for json mode
}

// FromJsonStr from json str to
func (r IDynamcHandler) FromJsonData(data map[string]interface{}, validate map[string]interface{}) []IDynamcHandler {
	res := []IDynamcHandler{}
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
			res = append(res, IDynamcHandler{
				Name:     key,
				Type:     "string",
				Mode:     JSON,
				Validate: curvalidatestr,
			})
		case int64, int32, int, uint64, uint32, uint:
			res = append(res, IDynamcHandler{
				Name:     key,
				Type:     "int",
				Mode:     JSON,
				Validate: curvalidatestr,
			})
		case float64, float32:
			res = append(res, IDynamcHandler{
				Name:     key,
				Type:     "float",
				Mode:     JSON,
				Validate: curvalidatestr,
			})
		case bool:
			res = append(res, IDynamcHandler{
				Name:     key,
				Type:     "bool",
				Mode:     JSON,
				Validate: curvalidatestr,
			})
		case map[string]interface{}:
			res = append(res, IDynamcHandler{
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
			res = append(res, IDynamcHandler{
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
				res = append(res, IDynamcHandler{
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

func (r IDynamcHandler) BuildModel(request []IDynamcHandler) (*Instance, string) {
	var contentType string

	mod := NewBuilder()
	for _, item := range request {
		var tag string

		validatestr := item.Validate
		nextvalidate := map[string]interface{}{}
		if err := json.Unmarshal([]byte(item.Validate), &nextvalidate); err != nil {
			logger.DefaultLogger.Warn(err.Error())
		}

		switch item.Mode {
		case Query:
			tag = fmt.Sprintf(`%s:"%s" validate:"%s"`, "query", item.Name, item.Validate)
		case Form:
			tag = fmt.Sprintf(`%s:"%s" validate:"%s"`, "form", item.Name, item.Validate)
			if contentType == "" {
				contentType = MIMEMultipartPOSTForm
			}
		case Header:
			tag = fmt.Sprintf(`%s:"%s" validate:"%s"`, "header", item.Name, item.Validate)
		case Uri:
			tag = fmt.Sprintf(`%s:"%s" validate:"%s"`, "uri", item.Name, item.Validate)
		case JSON:
			if len(nextvalidate) == 0 {
				// plain text
				tag = fmt.Sprintf(`%s:"%s" validate:"%s"`, "form", item.Name, validatestr)
			} else {
				// json text
				if v := nextvalidate["_"]; v != nil {
					if v, ok := v.(string); ok {
						tag = fmt.Sprintf(`%s:"%s" validate:"%s"`, "form", item.Name, v)
					}
				}
			}
			if contentType != "" {
				contentType = MIMEJSON
			}
		default:
			tag = fmt.Sprintf(`%s:"%s" validate:"%s"`, "form", item.Name, item.Validate)
		}

		switch item.Type {
		case "string":
			mod = mod.AddString(item.Name, tag)
		case "[]string":
			mod = mod.AddStringArray(item.Name, tag)
		case "int":
			mod = mod.AddInt64(item.Name, tag)
		case "[]int":
			mod = mod.AddInt64Array(item.Name, tag)
		case "float":
			mod = mod.AddFloat64(item.Name, tag)
		case "[]float":
			mod = mod.AddFloat64Array(item.Name, tag)
		case "bool":
			mod = mod.AddBool(item.Name, tag)
		case "[]bool":
			mod = mod.AddBoolArray(item.Name, tag)
		case "file":
			f := &multipart.FileHeader{}
			mod = mod.AddStruct(item.Name, f, tag, false)
		default:
			// maybe a json mod
			if len(item.children) > 0 {
				if item.Type == "array" {
					m, _ := r.BuildModel(item.children)
					mod = mod.AddStruct(item.Name, reflect.Zero(reflect.SliceOf(m.Type())).Interface(), tag, false)
				} else {
					m, _ := r.BuildModel(item.children)
					mod = mod.AddStruct(item.Name, reflect.New(m.Type()).Elem().Interface(), tag, false)
				}
				continue
			}

			var jsondata map[string]interface{}
			if err := json.Unmarshal([]byte(item.Type), &jsondata); err == nil {
				m, _ := r.BuildModel(r.FromJsonData(jsondata, nextvalidate))
				mod = mod.AddStruct(item.Name, reflect.New(m.Type()).Elem().Interface(), tag, false)
				continue
			}

			// todo the []map[string]interface{} type
			var jsondataArr []map[string]interface{}
			if err := json.Unmarshal([]byte(item.Type), &jsondataArr); err == nil && len(jsondataArr) == 1 {
				m, _ := r.BuildModel(r.FromJsonData(jsondataArr[0], nextvalidate))
				mod = mod.AddStruct(item.Name, reflect.Zero(reflect.SliceOf(m.Type())).Interface(), tag, false)
				continue
			}

			logger.DefaultLogger.Warn("不支持该数据类型")
		}
	}
	return mod.Build().New(), contentType
}

func (r IDynamcHandler) BindValue(request []IDynamcHandler, getVal func(name string) (interface{}, error)) (*Instance, error) {
	res, _ := r.BuildModel(request)
	var errs error
	for _, item := range request {
		val, err := getVal(item.Name)
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
			res.SetString(item.Name, fmt.Sprint(val))
		case "[]string":
			if cv, ok := val.([]string); ok {
				res.SetValue(item.Name, cv)
				continue
			}
			logger.DefaultLogger.Warn("not a []string type")
		case "int":
			if cv, err := strconv.ParseInt(fmt.Sprint(val), 10, 64); err != nil {
				logger.DefaultLogger.Warn("not a int64")
			} else {
				res.SetInt64(item.Name, cv)
			}
		case "[]int":
			if cv, ok := val.([]int64); ok {
				res.SetValue(item.Name, cv)
				continue
			}
			if cv, ok := val.([]int32); ok {
				res.SetValue(item.Name, cv)
				continue
			}
			if cv, ok := val.([]int); ok {
				res.SetValue(item.Name, cv)
				continue
			}
			logger.DefaultLogger.Warn("not a []string type")
		case "float":
			if cv, ok := val.(float64); !ok {
				logger.DefaultLogger.Warn("not a float")
				res.SetFloat64(item.Name, 0)
			} else {
				res.SetFloat64(item.Name, cv)
			}
		case "[]float":
			if cv, ok := val.([]float64); ok {
				res.SetValue(item.Name, cv)
				continue
			}
			logger.DefaultLogger.Warn("not a []float type")
		case "bool":
			if cv, ok := val.(bool); !ok {
				logger.DefaultLogger.Warn("not a bool")
				res.SetBool(item.Name, false)
			} else {
				res.SetBool(item.Name, cv)
			}
		case "[]bool":
			if cv, ok := val.([]bool); ok {
				res.SetValue(item.Name, cv)
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
		}
	}

	return res, nil
}
