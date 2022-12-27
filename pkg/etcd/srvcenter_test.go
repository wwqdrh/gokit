package etcd

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestSrvCenter(t *testing.T) {
	driver, err := NewEtcdDriver(&EtcdOptions{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
		UserName:    "root",
		Password:    "123456",
	})
	if err != nil {
		t.Skip(err.Error())
	}

	srv, err := NewSrvCenter(driver)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// 服务发现
	// 必须放在前面
	srcChan1 := srv.Discover(ctx, "service1")
	srcChan2 := srv.Discover(ctx, "service2")

	time.Sleep(1 * time.Second)

	// 注册服务一以及心跳
	srv1, err := srv.Distribution("service1", "8080", 5)
	if err != nil {
		t.Error(err)
	}
	go srv.BeatHeart(ctx, srv1, 3*time.Second)

	// 注册服务二以及心跳
	_, err = srv.Distribution("service2", "8081", 5)
	if err != nil {
		t.Error(err)
	}
	// srv2不续租，过期判断监听的位置是否会失效
	// go srv.BeatHeart(ctx, srv2, 5*time.Second)

	go func() {
		for item := range srcChan1 {
			fmt.Println(string(item.Kv.Key), string(item.Kv.Value))
			if string(item.Kv.Value) == "" {
				break
			}
		}
		fmt.Println("service1下线了")
	}()
	go func() {
		for item := range srcChan2 {
			fmt.Println(string(item.Kv.Key), string(item.Kv.Value))
			if string(item.Kv.Value) == "" {
				break
			}
		}
		fmt.Println("service2下线了")
	}()

	time.Sleep(5 * time.Second)
	cancel()
	time.Sleep(5 * time.Second)
}
