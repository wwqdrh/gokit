package rabbitmq

import (
	"context"
	"testing"
	"time"
)

func TestPubSub(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()

	pubqueue()
	subqueue(ctx)
}
