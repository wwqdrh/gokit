package client

import (
	"fmt"
	"testing"
)

func TestClientConfig(t *testing.T) {
	host, _, err := clientConfig("ssh://root:123456@127.0.0.1:22")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(host)
}
