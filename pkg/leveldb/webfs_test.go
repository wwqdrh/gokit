package render

import (
	"errors"
	"io"
	"io/fs"
	"testing"
)

func TestWebFS(t *testing.T) {
	var f fs.FS = NewWebFs(`https://unpkg.com/@wwqdrh/uikit@0.2.0/umd/assets`)
	data, err := f.Open("/images/tech/cover.png")
	if err != nil {
		t.Error(err)
		return
	}

	content, err := io.ReadAll(data)
	if err != nil {
		t.Error(err)
		return
	}
	if len(content) == 0 {
		t.Error(errors.New("读取内容失败"))
		return
	}
}
