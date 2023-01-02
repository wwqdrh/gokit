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

	if err := RegisterNotify(ctx, "./testdata", func(e fsnotify.Event) {
		fmt.Println("watch event")
	}); err != nil {
		fmt.Println(err.Error())
		return
	}

	if err := os.WriteFile("./testdata/temp.txt", []byte("1"), 0o755); err != nil {
		fmt.Println(err.Error())
		return
	}

	time.Sleep(2 * time.Second)

	if err := os.WriteFile("./testdata/a/temp.txt", []byte("1"), 0o755); err != nil {
		fmt.Println(err.Error())
		return
	}

	time.Sleep(2 * time.Second)

	// output:
	// watch event
	// watch event
}
