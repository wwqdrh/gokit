package logger

import (
	"bytes"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

// 包含两种属性: 颜色、输出格式
// 输出位置不是互斥，是可以组合的(由zapcore构造的时候指定)
// \033[1;31;40m    <!--1-高亮显示 31-前景色红色  40-背景色黑色-->
// \033[0m          <!--采用终端默认设置，即取消颜色设置-->
var (
	basic  = []byte("\033[0m")
	blue   = []byte("\033[36m")
	red    = []byte("\033[31m")
	yellow = []byte("\033[33m")
	green  = []byte("\033[34m")
)

// encoder

type ColorJsonEncoder struct {
	zapcore.Encoder

	config zapcore.EncoderConfig
}

type ColorConsoleEncoder struct {
	zapcore.Encoder
}

func withColorRender(level zapcore.Level, buf *buffer.Buffer) *buffer.Buffer {
	buffer := new(bytes.Buffer)
	if level >= zap.ErrorLevel {
		buffer.Write(red)
		buffer.Write(buf.Bytes())
		buffer.Write(basic)
	} else if level >= zap.WarnLevel {
		buffer.Write(yellow)
		buffer.Write(buf.Bytes())
		buffer.Write(basic)
	} else if level >= zap.InfoLevel {
		buffer.Write(blue)
		buffer.Write(buf.Bytes())
		buffer.Write(basic)
	} else {
		buffer.Write(green)
		buffer.Write(buf.Bytes())
		buffer.Write(basic)
	}

	buf.Reset()
	_, err := buf.Write(buffer.Bytes())
	if err != nil {
		return nil
	}
	return buf
}

func NewColorJsonEncoder(config zapcore.EncoderConfig) zapcore.Encoder {
	return ColorJsonEncoder{
		Encoder: zapcore.NewJSONEncoder(config),
		config:  config,
	}
}

func (enc ColorJsonEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf, err := enc.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return buf, err
	}

	return withColorRender(entry.Level, buf), nil
}

func NewColorConsoleEncoder(config zapcore.EncoderConfig) zapcore.Encoder {
	return ColorJsonEncoder{
		Encoder: zapcore.NewConsoleEncoder(config),
		config:  config,
	}
}

func (enc ColorConsoleEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf, err := enc.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return buf, err
	}

	return withColorRender(entry.Level, buf), nil
}
