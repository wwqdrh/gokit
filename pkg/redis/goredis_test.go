package redis

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RedisSuite struct {
	suite.Suite

	client *RedisDriver
}

func TestRedisSuite(t *testing.T) {
	d, err := NewRedisDriver(NewOption(WithPassword("123456")))
	if err != nil {
		fmt.Println(err.Error())
		t.Skip(err)
	}

	suite.Run(t, &RedisSuite{
		client: d,
	})
}

func (s *RedisSuite) TestLock() {
	var cnt int64 = 0
	wait := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			ok, err := s.client.Lock("httptool_test_lock_key")
			assert.Nil(s.T(), err)
			if ok {
				atomic.AddInt64(&cnt, 1)
			}
		}()
	}
	wait.Wait()
	assert.Equal(s.T(), int64(1), cnt)

	res, err := s.client.UnLock("httptool_test_lock_key")
	assert.Nil(s.T(), err)

	fmt.Println(res)
}
