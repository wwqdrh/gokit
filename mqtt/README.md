# MQTT Library

一个基于Eclipse Paho MQTT Go客户端的MQTT库，提供简单易用的API，支持连接、发布、订阅等核心功能。

## 功能特性

- 简单易用的API接口
- 支持上下文控制（context）
- 支持各种MQTT连接配置选项
- 支持发布消息
- 支持订阅消息
- 支持自动重连
- 支持优雅关闭

## 安装

使用Go模块安装：

```bash
go get github.com/wwqdrh/gokit/mqtt
```

## 基本使用

### 1. 连接到MQTT代理

```go
import (
	"context"
	"time"

	"github.com/wwqdrh/gokit/mqtt"
)

// 配置MQTT客户端
opts := mqtt.Options{
	Broker:               "tcp://broker.emqx.io:1883", // MQTT代理地址
	ClientID:             "your-client-id",
	Username:             "your-username",            // 可选
	Password:             "your-password",            // 可选
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
if err := client.Connect(ctx); err != nil {
	// 处理错误
}
```

### 2. 发布消息

```go
// 发布消息
if err := client.Publish(ctx, "test/topic", 1, false, "Hello MQTT!"); err != nil {
	// 处理错误
}
```

### 3. 订阅消息

```go
// 定义消息处理函数	handler := func(topic string, payload []byte) {
	fmt.Printf("Received message on topic '%s': %s\n", topic, string(payload))
}

// 订阅主题
if err := client.Subscribe(ctx, "test/topic", 1, handler); err != nil {
	// 处理错误
}
```

### 4. 取消订阅

```go
// 取消订阅
if err := client.Unsubscribe(ctx, "test/topic"); err != nil {
	// 处理错误
}
```

### 5. 断开连接

```go
// 断开连接
client.Disconnect(5 * time.Second)
```

## 配置选项

| 选项 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| Broker | string | MQTT代理地址，格式：tcp://host:port | - |
| ClientID | string | 客户端ID，应唯一 | - |
| Username | string | 连接用户名 | "" |
| Password | string | 连接密码 | "" |
| CleanSession | bool | 是否清除会话 | true |
| AutoReconnect | bool | 是否自动重连 | true |
| MaxReconnectInterval | time.Duration | 最大重连间隔 | 30s |
| KeepAlive | time.Duration | 保活间隔 | 60s |
| PingTimeout | time.Duration | Ping超时时间 | 10s |

## 示例代码

本项目提供了两个示例代码：

### 发布消息示例

```bash
cd examples
go run publisher.go
```

### 订阅消息示例

```bash
cd examples
go run subscriber.go
```

## 作为库引用

在其他项目中引用此库：

1. 在`go.mod`文件中添加依赖：

```go
require (
	github.com/wwqdrh/gokit/mqtt v0.0.0-xxxxxx
)
```

2. 导入并使用：

```go
import (
	"github.com/wwqdrh/gokit/mqtt"
)

// 使用方法与上述示例相同
```

## 依赖

- [github.com/eclipse/paho.mqtt.golang](https://github.com/eclipse/paho.mqtt.golang) - Eclipse Paho MQTT Go客户端

## 注意事项

- 示例代码使用公共MQTT代理`broker.emqx.io`，仅用于测试目的
- 在生产环境中，建议使用自己的MQTT代理
- 确保设置唯一的ClientID以避免连接冲突

## 自建服务器

```go
package mqtt

import (
	"context"
	"fmt"
	"log"

	mosquitto "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

// Server represents an MQTT server

type Server struct {
	mosquitto *mosquitto.Server
	address   string
}

// NewServer creates a new MQTT server instance
func NewServer(address string) *Server {
	s := &Server{
		mosquitto: mosquitto.New(nil),
		address:   address,
	}

	// Add allow-all authentication hook
	allowHook := &auth.AllowHook{}
	if err := s.mosquitto.AddHook(allowHook, nil); err != nil {
		log.Printf("Warning: Failed to add allow-all hook: %v", err)
	}

	return s
}

// Start starts the MQTT server
func (s *Server) Start() error {
	// Create MQTT listener config
	config := listeners.Config{
		ID:      "tcp",
		Type:    "tcp",
		Address: s.address,
	}

	// Create MQTT listener
	tcpListener := listeners.NewTCP(config)

	// Add listener to server
	if err := s.mosquitto.AddListener(tcpListener); err != nil {
		return fmt.Errorf("failed to add listener: %w", err)
	}

	// Start server
	go func() {
		if err := s.mosquitto.Serve(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	log.Printf("MQTT server started on %s", s.address)
	return nil
}

// Stop stops the MQTT server
func (s *Server) Stop(ctx context.Context) error {
	log.Println("Stopping MQTT server...")
	return s.mosquitto.Close()
}
```