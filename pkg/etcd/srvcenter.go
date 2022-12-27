package etcd

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// 基于etcd lease 租约 机制，每隔一段时间上传心跳续租

type SrvCenterHandler struct {
	driver *EtcdDriver
}

// 后台服务上报状态
func NewSrvCenter(driver *EtcdDriver) (*SrvCenterHandler, error) {
	return &SrvCenterHandler{
		driver: driver,
	}, nil
}

// 服务注册
func (s *SrvCenterHandler) Distribution(serviceName, serviceHost string, ttl int64) (clientv3.LeaseID, error) {
	resID, err := s.driver.PutWithLease(serviceName, serviceHost, ttl)
	if err != nil {
		return -1, err
	}
	return resID, nil
}

// 心跳上传状态
func (s *SrvCenterHandler) BeatHeart(ctx context.Context, leaseID clientv3.LeaseID, heart time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			time.Sleep(heart)
			if _, err := s.driver.KeepAlive(leaseID); err != nil {
				return err
			}
		}
	}
}

// 服务发现
// 基于监听
// 监听某个key对应的value
func (s *SrvCenterHandler) Discover(ctx context.Context, key string) chan *clientv3.Event {
	return s.driver.Watch(ctx, key)
}
