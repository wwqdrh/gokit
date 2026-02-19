package mqtt

import (
	"context"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	client mqtt.Client
	options *mqtt.ClientOptions
}

type Options struct {
	Broker   string
	ClientID string
	Username string
	Password string
	CleanSession bool
	AutoReconnect bool
	MaxReconnectInterval time.Duration
	KeepAlive time.Duration
	PingTimeout time.Duration
}

type MessageHandler func(topic string, payload []byte)

func NewClient(opts Options) *Client {
	clientOpts := mqtt.NewClientOptions()
	clientOpts.AddBroker(opts.Broker)
	clientOpts.SetClientID(opts.ClientID)
	clientOpts.SetUsername(opts.Username)
	clientOpts.SetPassword(opts.Password)
	clientOpts.SetCleanSession(opts.CleanSession)
	clientOpts.SetAutoReconnect(opts.AutoReconnect)
	clientOpts.SetMaxReconnectInterval(opts.MaxReconnectInterval)
	clientOpts.SetKeepAlive(opts.KeepAlive)
	clientOpts.SetPingTimeout(opts.PingTimeout)

	return &Client{
		options: clientOpts,
	}
}

func (c *Client) Connect(ctx context.Context) error {
	if c.client != nil && c.client.IsConnected() {
		return nil
	}

	c.client = mqtt.NewClient(c.options)
	token := c.client.Connect()

	select {
	case <-token.Done():
		return token.Error()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) Publish(ctx context.Context, topic string, qos byte, retained bool, payload interface{}) error {
	if c.client == nil || !c.client.IsConnected() {
		if err := c.Connect(ctx); err != nil {
			return err
		}
	}

	token := c.client.Publish(topic, qos, retained, payload)

	select {
	case <-token.Done():
		return token.Error()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) Subscribe(ctx context.Context, topic string, qos byte, handler MessageHandler) error {
	if c.client == nil || !c.client.IsConnected() {
		if err := c.Connect(ctx); err != nil {
			return err
		}
	}

	mqttHandler := func(client mqtt.Client, msg mqtt.Message) {
		handler(msg.Topic(), msg.Payload())
	}

	token := c.client.Subscribe(topic, qos, mqttHandler)

	select {
	case <-token.Done():
		return token.Error()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) Unsubscribe(ctx context.Context, topic string) error {
	if c.client == nil || !c.client.IsConnected() {
		return nil
	}

	token := c.client.Unsubscribe(topic)

	select {
	case <-token.Done():
		return token.Error()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) Disconnect(waitTime time.Duration) {
	if c.client != nil {
		c.client.Disconnect(uint(waitTime.Milliseconds()))
	}
}

func (c *Client) IsConnected() bool {
	if c.client == nil {
		return false
	}
	return c.client.IsConnected()
}
