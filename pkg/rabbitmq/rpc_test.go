package rabbitmq

import (
	"context"
	"testing"
	"time"
)

func TestRpc(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()
	go func() {
		rpcserver(ctx)
	}()

	time.Sleep(1 * time.Second)
	rpcclient()
	<-ctx.Done()
}
