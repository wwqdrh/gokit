package main

import (
	// "errors"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

var (
	ErrNameEmpty = errors.New("Name cant't be empty!")
	ErrUnknown   = errors.New("unknonw")
)

type Student struct {
	Name string
	Age  int
}

func NewStu() (err error) {
	stu := &Student{
		Age: 19,
	}
	e := stu.SetName("")
	if e != nil {
		return errors.Wrap(e, "set name failed!")
	} else {
		return nil
	}
}

func (s *Student) SetName(newName string) (err error) {
	if newName == "" || s.Name == "" {
		return errors.WithStack(ErrNameEmpty)
	} else {
		s.Name = newName
		return nil
	}
}

// GetErrorStack function
func GetErrorStack(err error) []string {
	// initialize an empty slice to store the error stack
	stack := []string{}
	// convert the error to a string and split it by newline characters
	errStr := fmt.Sprintf("%+v", err)
	lines := strings.Split(errStr, "\n")
	// loop through the lines and find the ones that start with "\t"
	for _, line := range lines {
		if strings.HasPrefix(line, "\t") {
			// trim the leading and trailing spaces and append it to the stack slice
			line = strings.TrimSpace(line)
			stack = append(stack, line)
		}
	}
	// return the stack slice
	return stack
}

func main() {
	e := NewStu()
	fmt.Println(e.Error())
	fmt.Printf("%+v\n", e)
	fmt.Println(GetErrorStack(e))
	fmt.Println(errors.Is(e, ErrNameEmpty))
}
