package cache

import (
	"context"
	"time"
)

type Cache interface {
	GetDefaultExpiration() time.Duration
	Set(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string, to interface{}) (exist bool, err error)
	Delete(ctx context.Context, key string) error
	SetBytes(ctx context.Context, key string, value []byte) error
	GetBytes(ctx context.Context, key string) ([]byte, error)
}
