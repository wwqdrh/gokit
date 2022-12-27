package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisDriver struct {
	client *redis.Client

	opt *option
}

func NewRedisDriver(opt *option) (*RedisDriver, error) {
	driver := &RedisDriver{
		client: redis.NewClient(&redis.Options{
			Addr:     opt.addr,
			Password: opt.password,
			DB:       opt.db, // default DB,
			// TLSConfig: &tls.Config{},
		}),
		opt: opt,
	}
	if err := driver.Ping(); err != nil {
		return nil, err
	}
	return driver, nil
}

// 检测连接状态
func (d *RedisDriver) Ping() error {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	_, err := d.client.Ping(ctx).Result()

	return err
}

// redis分布式锁

func (d *RedisDriver) Lock(key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	return d.client.SetNX(ctx, key, 1, 10*time.Second).Result()
}

// 为了避免其他工作任务解到锁，可以使用一个唯一value进行标识
func (d *RedisDriver) UnLock(key string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	nums, err := d.client.Del(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return nums, nil
}

////////////////////
// pipeline
////////////////////

// Pipeline 主要是一种网络优化。它本质上意味着客户端缓冲一堆命令并一次性将它们发送到服务器。
// 这些命令不能保证在事务中执行。
// 这样做的好处是节省了每个命令的网络往返时间（RTT）。

func (d *RedisDriver) PipelineIncr() {
	ctx := context.Background()

	// 1、
	var incr *redis.IntCmd
	_, err := d.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		incr = pipe.Incr(ctx, "pipelined_counter")
		pipe.Expire(ctx, "pipelined_counter", time.Second*5)
		return nil
	})
	if err != nil {
		fmt.Printf("pipeline_counter err: %v", err)
	} else {
		fmt.Println(incr.Val())
	}

	// 2、
	pipe := d.client.Pipeline()
	incr = pipe.Incr(ctx, "pipeline_counter2")
	pipe.Expire(ctx, "pipeline_counter2", time.Second*5)
	_, err = pipe.Exec(ctx)
	if err != nil {
		fmt.Printf("pipeline_counter2 err: %v", err)
	} else {
		fmt.Println(incr.Val())
	}
}

////////////////////
// pubsub
////////////////////

func (d *RedisDriver) TryPub(channel, message string) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	err := d.client.Publish(ctx, channel, message).Err()
	if err != nil {
		fmt.Println("发生错误")
	}
}

func (d *RedisDriver) TrySub(channel string) string {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	sub := d.client.Subscribe(ctx, channel)
	defer sub.Close()

	// for {
	// 	msg, err := sub.ReceiveMessage(ctx)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Println(msg.Channel, msg.Payload)
	// }
	ch := sub.Channel()
	var message string
	for msg := range ch {
		fmt.Println(msg.Channel, msg.Payload)
		message = msg.Payload
	}
	return message
}

////////////////////
// transaction
////////////////////

func (d *RedisDriver) Transaction(key string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	maxRetries := 1000
	txf := func(tx *redis.Tx) error {
		n, err := tx.Get(ctx, key).Int() // key不存在的时候值为0
		if err != nil && err != redis.Nil {
			return err
		}
		n++
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, key, n, 0)
			return nil
		})
		return err
	}
	for i := 0; i < maxRetries; i++ {
		err := d.client.Watch(ctx, txf, key)
		if err == nil {
			return nil
		}
		if err == redis.TxFailedErr {
			continue
		}
		return err
	}
	return errors.New("increment reached maximum number of retries")
}

////////////////////
// crud
////////////////////

// string
func (d *RedisDriver) StringGet(key string) string {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	val, err := d.client.Get(ctx, key).Result()
	switch {
	case err == redis.Nil:
		fmt.Println("key dose not exist")
	case err != nil:
		fmt.Println("Get failed", err)
	case val == "":
		fmt.Println("value is empty")
	}
	return val
}

func (d *RedisDriver) StringSet(key, value string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := d.client.Set(ctx, key, value, 5*time.Second).Err()
	if err != nil {
		fmt.Printf("set failed, err:%v\n", err)
		return false
	}
	return true
}

func (d *RedisDriver) StringDel(key string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := d.client.Del(ctx, key).Result()
	if err != nil {
		fmt.Println("redis删除失败")
		return false
	}
	return true
}
