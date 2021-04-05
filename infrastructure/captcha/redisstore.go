package captcha

import (
	"context"
	"github.com/dchest/captcha"
	"github.com/huskar-t/gopher/common/define/cache"
	"time"
)

type redisStore struct {
	timeout time.Duration
	cache   cache.Cache
	ctx     context.Context
}

func NewStore(cache cache.Cache) captcha.Store {
	return &redisStore{
		timeout: cache.GetDefaultExpiration(),
		cache:   cache,
		ctx:     context.Background(),
	}
}

// Set sets the digits for the captcha id.
func (s *redisStore) Set(id string, digits []byte) {
	_ = s.cache.SetBytes(s.ctx, id, digits)
}

// Get returns stored digits for the captcha id. Clear indicates
// whether the captcha must be deleted from the store.
func (s *redisStore) Get(id string, clear bool) (digits []byte) {
	result, err := s.cache.GetBytes(s.ctx, id)
	if err != nil {
		return
	}
	if clear {
		s.cache.Delete(s.ctx, id)
	}
	return result
}
