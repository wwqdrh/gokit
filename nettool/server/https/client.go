package https

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"

	"github.com/tidwall/gjson"
)

type HTTPOpt struct {
	Engine    http.Handler
	Handler   http.HandlerFunc
	Method    string
	Url       string
	Header    map[string]string
	Cookies   []*http.Cookie
	Body      io.Reader
	Query     map[string][]string
	JSON      map[string]interface{}
	MultiForm map[string]Field

	contentType string // write by data build
}

type Field struct {
	Value  string
	IsFile bool
}

type HTTPRes struct {
	Code    int
	Header  map[string][]string
	Body    string
	Cookies []*http.Cookie
}

func DoReq(opt *HTTPOpt) (HTTPRes, error) {
	if opt.Engine == nil {
		if opt.Handler == nil {
			return HTTPRes{}, errors.New("cant engine handler both nil")
		}
		if opt.Url == "" {
			opt.Url = "/"
		}
		mutex := http.NewServeMux()
		mutex.HandleFunc(opt.Url, opt.Handler)
		opt.Engine = mutex
	}

	urlpath, body := opt.Url, opt.Body

	var err error
	if len(opt.JSON) > 0 {
		body, err = opt.jsonBuild(opt.JSON)
		if err != nil {
			return HTTPRes{}, err
		}
	} else if len(opt.MultiForm) > 0 {
		body, err = opt.multiFormBuild(opt.MultiForm)
		if err != nil {
			return HTTPRes{}, err
		}
	}

	if len(opt.Query) > 0 {
		urlpath = urlpath + "?" + url.Values(opt.Query).Encode()
	}

	req, res := httptest.NewRequest(opt.Method, urlpath, body), httptest.NewRecorder()
	for key, val := range opt.Header {
		req.Header.Add(key, val)
	}
	req.Header.Add("Content-Type", opt.contentType)
	for _, cookie := range opt.Cookies {
		req.AddCookie(cookie)
	}

	opt.Engine.ServeHTTP(res, req)

	return HTTPRes{
		Code:    res.Code,
		Header:  res.Header(),
		Body:    res.Body.String(),
		Cookies: res.Result().Cookies(),
	}, nil
}

func (o *HTTPOpt) multiFormBuild(f map[string]Field) (io.Reader, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, val := range f {
		if !val.IsFile {
			if err := writer.WriteField(key, val.Value); err != nil {
				return nil, err
			}
		} else {
			part, err := writer.CreateFormFile(key, filepath.Base(val.Value))
			if err != nil {
				return nil, err
			}
			file, err := os.Open(val.Value)
			if err != nil {
				return nil, err
			}
			_, err = io.Copy(part, file)
			if err != nil {
				fmt.Println(err.Error())
			}
			file.Close()
		}
	}
	writer.Close() // 必须写在在这，否则服务端会报没有EOF
	o.contentType = writer.FormDataContentType()
	return body, nil
}

func (o *HTTPOpt) jsonBuild(f map[string]interface{}) (io.Reader, error) {
	bodyData, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(bodyData)
	o.contentType = "application/json"
	return body, nil
}

func (r *HTTPRes) GJson(key string) gjson.Result {
	return gjson.Get(r.Body, key)
}
