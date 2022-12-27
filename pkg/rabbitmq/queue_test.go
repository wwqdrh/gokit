package rabbitmq

import (
	"context"
	"testing"
	"time"
)

func TestWorkQueue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()

	queueNewTask()
	queueNewWorker(ctx)
}
