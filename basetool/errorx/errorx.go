package gerror

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
)

// Option is option for creating error.
type Option struct {
	Error error     // Wrapped error if any.
	Stack bool      // Whether recording stack information into error.
	Text  string    // Error text, which is created by New* functions.
	Code  ErrorCode // Error code if necessary.
}

// NewOption creates and returns an error with Option.
// It is the senior usage for creating error, which is often used internally in framework.
func NewOption(option Option) error {
	err := &Error{
		error: option.Error,
		text:  option.Text,
		code:  option.Code,
	}
	if option.Stack {
		err.stack = callers()
	}
	return err
}

// stack represents a stack of program counters.
type stack []uintptr

const (
	// maxStackDepth marks the max stack depth for error back traces.
	maxStackDepth = 32
)

// callers returns the stack callers.
// Note that it here just retrieves the caller memory address array not the caller information.
func callers(skip ...int) stack {
	var (
		pcs [maxStackDepth]uintptr
		n   = 3
	)
	if len(skip) > 0 {
		n += skip[0]
	}
	return pcs[:runtime.Callers(n, pcs[:])]
}

// iCode is the interface for Code feature.
type iCode interface {
	Error() string
	Code() ErrorCode
}

// iStack is the interface for Stack feature.
type iStack interface {
	Error() string
	Stack() string
}

// iCause is the interface for Cause feature.
type iCause interface {
	Error() string
	Cause() error
}

// iCurrent is the interface for Current feature.
type iCurrent interface {
	Error() string
	Current() error
}

// iNext is the interface for Next feature.
type iNext interface {
	Error() string
	Next() error
}

// New creates and returns an error which is formatted from given text.
func ErrorNew(text string) error {
	return &Error{
		stack: callers(),
		text:  text,
		code:  CodeNil,
	}
}

// Newf returns an error that formats as the given format and args.
func Newf(format string, args ...interface{}) error {
	return &Error{
		stack: callers(),
		text:  fmt.Sprintf(format, args...),
		code:  CodeNil,
	}
}

// NewSkip creates and returns an error which is formatted from given text.
// The parameter `skip` specifies the stack callers skipped amount.
func NewSkip(skip int, text string) error {
	return &Error{
		stack: callers(skip),
		text:  text,
		code:  CodeNil,
	}
}

// NewSkipf returns an error that formats as the given format and args.
// The parameter `skip` specifies the stack callers skipped amount.
func NewSkipf(skip int, format string, args ...interface{}) error {
	return &Error{
		stack: callers(skip),
		text:  fmt.Sprintf(format, args...),
		code:  CodeNil,
	}
}

// Wrap wraps error with text. It returns nil if given err is nil.
// Note that it does not lose the error code of wrapped error, as it inherits the error code from it.
func Wrap(err error, text string) error {
	if err == nil {
		return nil
	}
	return &Error{
		error: err,
		stack: callers(),
		text:  text,
		code:  Code(err),
	}
}

// Wrapf returns an error annotating err with a stack trace at the point Wrapf is called, and the format specifier.
// It returns nil if given `err` is nil.
// Note that it does not lose the error code of wrapped error, as it inherits the error code from it.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &Error{
		error: err,
		stack: callers(),
		text:  fmt.Sprintf(format, args...),
		code:  Code(err),
	}
}

// WrapSkip wraps error with text. It returns nil if given err is nil.
// The parameter `skip` specifies the stack callers skipped amount.
// Note that it does not lose the error code of wrapped error, as it inherits the error code from it.
func WrapSkip(skip int, err error, text string) error {
	if err == nil {
		return nil
	}
	return &Error{
		error: err,
		stack: callers(skip),
		text:  text,
		code:  Code(err),
	}
}

// WrapSkipf wraps error with text that is formatted with given format and args. It returns nil if given err is nil.
// The parameter `skip` specifies the stack callers skipped amount.
// Note that it does not lose the error code of wrapped error, as it inherits the error code from it.
func WrapSkipf(skip int, err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &Error{
		error: err,
		stack: callers(skip),
		text:  fmt.Sprintf(format, args...),
		code:  Code(err),
	}
}

// NewCode creates and returns an error that has error code and given text.
func NewCode(code ErrorCode, text ...string) error {
	return &Error{
		stack: callers(),
		text:  strings.Join(text, ", "),
		code:  code,
	}
}

// NewCodef returns an error that has error code and formats as the given format and args.
func NewCodef(code ErrorCode, format string, args ...interface{}) error {
	return &Error{
		stack: callers(),
		text:  fmt.Sprintf(format, args...),
		code:  code,
	}
}

// NewCodeSkip creates and returns an error which has error code and is formatted from given text.
// The parameter `skip` specifies the stack callers skipped amount.
func NewCodeSkip(code ErrorCode, skip int, text ...string) error {
	return &Error{
		stack: callers(skip),
		text:  strings.Join(text, ", "),
		code:  code,
	}
}

// NewCodeSkipf returns an error that has error code and formats as the given format and args.
// The parameter `skip` specifies the stack callers skipped amount.
func NewCodeSkipf(code ErrorCode, skip int, format string, args ...interface{}) error {
	return &Error{
		stack: callers(skip),
		text:  fmt.Sprintf(format, args...),
		code:  code,
	}
}

// WrapCode wraps error with code and text.
// It returns nil if given err is nil.
func WrapCode(code ErrorCode, err error, text ...string) error {
	if err == nil {
		return nil
	}
	return &Error{
		error: err,
		stack: callers(),
		text:  strings.Join(text, ", "),
		code:  code,
	}
}

// WrapCodef wraps error with code and format specifier.
// It returns nil if given `err` is nil.
func WrapCodef(code ErrorCode, err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &Error{
		error: err,
		stack: callers(),
		text:  fmt.Sprintf(format, args...),
		code:  code,
	}
}

// WrapCodeSkip wraps error with code and text.
// It returns nil if given err is nil.
// The parameter `skip` specifies the stack callers skipped amount.
func WrapCodeSkip(code ErrorCode, skip int, err error, text ...string) error {
	if err == nil {
		return nil
	}
	return &Error{
		error: err,
		stack: callers(skip),
		text:  strings.Join(text, ", "),
		code:  code,
	}
}

// WrapCodeSkipf wraps error with code and text that is formatted with given format and args.
// It returns nil if given err is nil.
// The parameter `skip` specifies the stack callers skipped amount.
func WrapCodeSkipf(code ErrorCode, skip int, err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &Error{
		error: err,
		stack: callers(skip),
		text:  fmt.Sprintf(format, args...),
		code:  code,
	}
}

// Code returns the error code of current error.
// It returns CodeNil if it has no error code neither it does not implement interface Code.
func Code(err error) ErrorCode {
	if err == nil {
		return CodeNil
	}
	if e, ok := err.(iCode); ok {
		return e.Code()
	}
	if e, ok := err.(iNext); ok {
		return Code(e.Next())
	}
	return CodeNil
}

// Cause returns the root cause error of `err`.
func Cause(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(iCause); ok {
		return e.Cause()
	}
	if e, ok := err.(iNext); ok {
		return Cause(e.Next())
	}
	return err
}

// Stack returns the stack callers as string.
// It returns the error string directly if the `err` does not support stacks.
func Stack(err error) string {
	if err == nil {
		return ""
	}
	if e, ok := err.(iStack); ok {
		return e.Stack()
	}
	return err.Error()
}

// Current creates and returns the current level error.
// It returns nil if current level error is nil.
func Current(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(iCurrent); ok {
		return e.Current()
	}
	return err
}

// Next returns the next level error.
// It returns nil if current level error or the next level error is nil.
func Next(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(iNext); ok {
		return e.Next()
	}
	return nil
}

// HasStack checks and returns whether `err` implemented interface `iStack`.
func HasStack(err error) bool {
	_, ok := err.(iStack)
	return ok
}

// Error is custom error for additional features.
type Error struct {
	error error     // Wrapped error.
	stack stack     // Stack array, which records the stack information when this error is created or wrapped.
	text  string    // Custom Error text when Error is created, might be empty when its code is not nil.
	code  ErrorCode // Error code if necessary.
}

const (
	// Filtering key for current error module paths.
	stackFilterKeyLocal = "/errors/gerror/gerror"
)

var (
	// goRootForFilter is used for stack filtering in development environment purpose.
	goRootForFilter = runtime.GOROOT()
)

func init() {
	if goRootForFilter != "" {
		goRootForFilter = strings.Replace(goRootForFilter, "\\", "/", -1)
	}
}

// Error implements the interface of Error, it returns all the error as string.
func (err *Error) Error() string {
	if err == nil {
		return ""
	}
	errStr := err.text
	if errStr == "" && err.code != nil {
		errStr = err.code.Message()
	}
	if err.error != nil {
		if errStr != "" {
			errStr += ": "
		}
		errStr += err.error.Error()
	}
	return errStr
}

// Code returns the error code.
// It returns CodeNil if it has no error code.
func (err *Error) Code() ErrorCode {
	if err == nil {
		return CodeNil
	}
	if err.code == CodeNil {
		return Code(err.Next())
	}
	return err.code
}

// Cause returns the root cause error.
func (err *Error) Cause() error {
	if err == nil {
		return nil
	}
	loop := err
	for loop != nil {
		if loop.error != nil {
			if e, ok := loop.error.(*Error); ok {
				// Internal Error struct.
				loop = e
			} else if e, ok := loop.error.(iCause); ok {
				// Other Error that implements ApiCause interface.
				return e.Cause()
			} else {
				return loop.error
			}
		} else {
			// return loop
			// To be compatible with Case of https://github.com/pkg/errors.
			return errors.New(loop.text)
		}
	}
	return nil
}

// Format formats the frame according to the fmt.Formatter interface.
//
// %v, %s   : Print all the error string;
// %-v, %-s : Print current level error string;
// %+s      : Print full stack error list;
// %+v      : Print the error string and full stack error list;
func (err *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		switch {
		case s.Flag('-'):
			if err.text != "" {
				_, _ = io.WriteString(s, err.text)
			} else {
				_, _ = io.WriteString(s, err.Error())
			}
		case s.Flag('+'):
			if verb == 's' {
				_, _ = io.WriteString(s, err.Stack())
			} else {
				_, _ = io.WriteString(s, err.Error()+"\n"+err.Stack())
			}
		default:
			_, _ = io.WriteString(s, err.Error())
		}
	}
}

// Stack returns the stack callers as string.
// It returns an empty string if the `err` does not support stacks.
func (err *Error) Stack() string {
	if err == nil {
		return ""
	}
	var (
		loop   = err
		index  = 1
		buffer = bytes.NewBuffer(nil)
	)
	for loop != nil {
		buffer.WriteString(fmt.Sprintf("%d. %-v\n", index, loop))
		index++
		formatSubStack(loop.stack, buffer)
		if loop.error != nil {
			if e, ok := loop.error.(*Error); ok {
				loop = e
			} else {
				buffer.WriteString(fmt.Sprintf("%d. %s\n", index, loop.error.Error()))
				index++
				break
			}
		} else {
			break
		}
	}
	return buffer.String()
}

// Current creates and returns the current level error.
// It returns nil if current level error is nil.
func (err *Error) Current() error {
	if err == nil {
		return nil
	}
	return &Error{
		error: nil,
		stack: err.stack,
		text:  err.text,
		code:  err.code,
	}
}

// Next returns the next level error.
// It returns nil if current level error or the next level error is nil.
func (err *Error) Next() error {
	if err == nil {
		return nil
	}
	return err.error
}

// SetCode updates the internal code with given code.
func (err *Error) SetCode(code ErrorCode) {
	if err == nil {
		return
	}
	err.code = code
}

// MarshalJSON implements the interface MarshalJSON for json.Marshal.
// Note that do not use pointer as its receiver here.
func (err Error) MarshalJSON() ([]byte, error) {
	return []byte(`"` + err.Error() + `"`), nil
}

// formatSubStack formats the stack for error.
func formatSubStack(st stack, buffer *bytes.Buffer) {
	if st == nil {
		return
	}
	index := 1
	space := "  "
	for _, p := range st {
		if fn := runtime.FuncForPC(p - 1); fn != nil {
			file, line := fn.FileLine(p - 1)
			// Custom filtering.
			if strings.Contains(file, stackFilterKeyLocal) {
				continue
			}
			// Avoid stack string like "`autogenerated`"
			if strings.Contains(file, "<") {
				continue
			}
			// Ignore GO ROOT paths.
			if goRootForFilter != "" &&
				len(file) >= len(goRootForFilter) &&
				file[0:len(goRootForFilter)] == goRootForFilter {
				continue
			}
			// Graceful indent.
			if index > 9 {
				space = " "
			}
			buffer.WriteString(fmt.Sprintf(
				"   %d).%s%s\n    \t%s:%d\n",
				index, space, fn.Name(), file, line,
			))
			index++
		}
	}
}
