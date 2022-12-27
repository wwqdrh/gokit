package rabbitmq

import (
	"context"
	"testing"
	"time"
)

func TestBasicSendAndRecive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()

	BasicSend()
	BasicRecive(ctx)
}
