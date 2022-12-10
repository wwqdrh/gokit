package main

import (
	"context"
	"fmt"
	"time"

	"github.com/wwqdrh/gokit/logger/pluginx/pyroscopex"
)

func main() {
	pyroscopex.Start(context.Background(), pyroscopex.NewPprofOption(
		"http://127.0.0.1:4040",
		"simple.golang.app",
		pyroscopex.WithPprofType(pyroscopex.AllTypeOptions...),
	))

	for i := 0; i < 100; i++ {
		_ = make([]int, 10)
		time.Sleep(5 * time.Second)
	}
	fmt.Println("done")
}
