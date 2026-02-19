package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wwqdrh/gokit/mqtt"
)

func main() {
	// 配置MQTT客户端
	opts := mqtt.Options{
		Broker:               "broker.emqx.io:1883", // 使用本地MQTT服务器
		ClientID:             fmt.Sprintf("subscriber-%d", time.Now().Unix()),
		Username:             "",
		Password:             "",
		CleanSession:         true,
		AutoReconnect:        true,
		MaxReconnectInterval: 30 * time.Second,
		KeepAlive:            60 * time.Second,
		PingTimeout:          10 * time.Second,
	}

	// 创建MQTT客户端
	client := mqtt.NewClient(opts)

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 连接到MQTT代理
	fmt.Println("Connecting to MQTT broker...")
	if err := client.Connect(ctx); err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}
	fmt.Println("Connected successfully!")

	// 定义消息处理函数
	handler := func(topic string, payload []byte) {
		fmt.Printf("Received message on topic '%s': %s\n", topic, string(payload))
	}

	// 订阅主题
	topic := "test/topic"
	qos := byte(1)

	fmt.Printf("Subscribing to topic '%s'...\n", topic)
	if err := client.Subscribe(ctx, topic, qos, handler); err != nil {
		fmt.Printf("Failed to subscribe: %v\n", err)
		return
	}
	fmt.Printf("Subscribed to topic '%s' successfully!\n", topic)
	fmt.Println("Waiting for messages... (Press Ctrl+C to exit)")

	// 等待中断信号以优雅关闭
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// 取消订阅
	fmt.Printf("Unsubscribing from topic '%s'...\n", topic)
	if err := client.Unsubscribe(ctx, topic); err != nil {
		fmt.Printf("Failed to unsubscribe: %v\n", err)
	}
	fmt.Printf("Unsubscribed from topic '%s' successfully!\n", topic)

	// 断开连接
	fmt.Println("Disconnecting from MQTT broker...")
	client.Disconnect(5 * time.Second)
	fmt.Println("Disconnected successfully!")
}
