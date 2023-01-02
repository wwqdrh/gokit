package ostool

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

func ExampleRegisterNotify() {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	defer os.Remove("./testdata/temp.txt")
	defer os.Remove("./testdata/a/temp.txt")

	if err := RegisterNotify(ctx, "./testdata", 2*time.Second, func(e fsnotify.Event) {
		fmt.Println("watch event")
	}); err != nil {
		fmt.Println(err.Error())
		return
	}

	time.Sleep(1 * time.Second)

	if err := os.WriteFile("./testdata/temp.txt", []byte("1"), 0o755); err != nil {
		fmt.Println(err.Error())
		return
	}

	time.Sleep(2 * time.Second)

	if err := os.WriteFile("./testdata/a/temp.txt", []byte("1"), 0o755); err != nil {
		fmt.Println(err.Error())
		return
	}

	time.Sleep(1 * time.Second)

	// output:
	// watch event
	// watch event
	// watch event
	// watch event
}

func ExampleRegisterWhenDeleteDir() {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	if err := RegisterNotify(ctx, "./testdata", 2*time.Minute, func(e fsnotify.Event) {
		fmt.Println("watch event")
	}); err != nil {
		fmt.Println(err.Error())
		return
	}
	time.Sleep(2 * time.Second)

	// 创建子文件夹的事件
	if err := os.Mkdir("./testdata/a/b", 0o777); err != nil {
		fmt.Println(err.Error())
		return
	}
	defer os.RemoveAll("./testdata/a/b")
	time.Sleep(2 * time.Second)
	// 判断中途创建文件是否能捕捉到 will get write and create
	if err := os.WriteFile("./testdata/a/b/temp.txt", []byte("1"), 0o755); err != nil {
		fmt.Println(err.Error())
		return
	}
	time.Sleep(2 * time.Second)
	cancel()
	time.Sleep(2 * time.Second)

	// output:
	// watch event
	// watch event
}
