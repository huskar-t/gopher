package etcdlock

import (
	"context"
	"github.com/pkg/errors"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"time"
)

type EtcdMutex struct {
	s *concurrency.Session
	m *concurrency.Mutex
}

func NewMutex(key string, client *clientv3.Client) (mutex *EtcdMutex, err error) {
	mutex = &EtcdMutex{}
	mutex.s, err = concurrency.NewSession(client)
	if err != nil {
		return
	}
	mutex.m = concurrency.NewMutex(mutex.s, key)
	return
}

func (mutex *EtcdMutex) Lock() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //设置5s超时
	defer cancel()
	if err = mutex.m.Lock(ctx); err != nil {
		err = errors.Wrap(err, "获取分布式锁失败")
	}
	return
}
func (mutex *EtcdMutex) Unlock() (err error) {
	err = mutex.m.Unlock(context.TODO())

	if err != nil {
		return
	}
	err = mutex.s.Close()

	if err != nil {
		return
	}
	return
}
