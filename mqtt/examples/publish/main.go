package main

import (
	"context"
	"fmt"
	"time"

	"github.com/wwqdrh/gokit/mqtt"
)

func main() {
	// 配置MQTT客户端
	opts := mqtt.Options{
		Broker:               "broker.emqx.io:1883", // 使用本地MQTT服务器
		ClientID:             fmt.Sprintf("publisher-%d", time.Now().Unix()),
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

	// 发布消息
	topic := "/api/device1/ping"
	payload := fmt.Sprintf("Hello MQTT! Time: %s", time.Now().Format(time.RFC3339))
	qos := byte(1)
	retained := false

	fmt.Printf("Publishing message to topic '%s': %s\n", topic, payload)
	if err := client.Publish(ctx, topic, qos, retained, payload); err != nil {
		fmt.Printf("Failed to publish message: %v\n", err)
		return
	}
	fmt.Println("Message published successfully!")

	// 断开连接
	fmt.Println("Disconnecting from MQTT broker...")
	client.Disconnect(5 * time.Second)
	fmt.Println("Disconnected successfully!")
}
