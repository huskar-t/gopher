package redis

import (
	"bytes"
	"context"
	"encoding/gob"
	redisBase "github.com/go-redis/redis/v8"
	"time"
)

type Cache struct {
	conn              redisBase.UniversalClient
	group             string
	defaultExpiration time.Duration
}

func (c *Cache) key(key string) string {
	if c.group != "" {
		return c.group + ":" + key
	}
	return key
}
func (c *Cache) Set(ctx context.Context, key string, value interface{}) error {
	k := c.key(key)
	payload, err := encode(value)
	if err != nil {
		return err
	}
	return c.conn.Set(ctx, k, payload, c.defaultExpiration).Err()
}

func (c *Cache) Get(ctx context.Context, key string, to interface{}) (exist bool, err error) {
	k := c.key(key)
	payload, err := c.conn.Get(ctx, k).Bytes()
	if err != nil {
		if err == redisBase.Nil {
			return false, nil
		}
		return false, err
	}
	err = decode(payload, to)
	if err == nil {
		return true, nil
	}
	return true, err
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	k := c.key(key)
	return c.conn.Del(ctx, k).Err()
}
func (c *Cache) SetBytes(ctx context.Context, key string, value []byte) error {
	k := c.key(key)
	return c.conn.Set(ctx, k, value, c.defaultExpiration).Err()
}
func (c *Cache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	k := c.key(key)
	return c.conn.Get(ctx, k).Bytes()
}
func (c *Cache) GetDefaultExpiration() time.Duration {
	return c.defaultExpiration
}
func NewCache(client redisBase.UniversalClient, group string, defaultExpiration time.Duration) *Cache {
	return &Cache{
		group:             group,
		defaultExpiration: defaultExpiration,
		conn:              client,
	}
}

// Encode 用gob进行数据编码
func encode(data interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode 用gob进行数据解码
func decode(data []byte, to interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(to)
}
