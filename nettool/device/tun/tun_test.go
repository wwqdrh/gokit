package tun

import (
	"context"
	"testing"
	"time"
)

func TestToSocks(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	if err := Ins().CheckContext(); err != nil {
		t.Error(err)
	}
	if err := Ins().ToSocks(ctx, "127.0.0.1:2000"); err != nil {
		t.Error(err)
	}
}
