package rabbitmq

import (
	"context"
	"testing"
	"time"
)

func TestTopic(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()

	topicsend()
	topicrecive(ctx)
}
