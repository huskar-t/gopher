package redis

import (
	"context"
	redisBase "github.com/go-redis/redis/v8"
	"strings"
	"time"
)

type Config struct {
	Addr           string
	Password       string
	Cluster        bool
	DB             int
	MaxIdle        int
	MaxActive      int
	IdleTimeout    int // 240s
	ConnectTimeout int // 10s
	ReadTimeout    int // 10s
	WriteTimeout   int // 10s
}

func NewClient(conf *Config) (redisBase.UniversalClient, error) {
	if conf.IdleTimeout <= 0 {
		conf.IdleTimeout = 240
	}
	if conf.ConnectTimeout <= 0 {
		conf.ConnectTimeout = 10
	}
	if conf.ReadTimeout <= 0 {
		conf.ReadTimeout = 10
	}
	if conf.WriteTimeout <= 0 {
		conf.WriteTimeout = 10
	}
	idleTimeout := time.Duration(conf.IdleTimeout) * time.Second
	connectTimeout := time.Duration(conf.ConnectTimeout) * time.Second
	readTimeout := time.Duration(conf.ReadTimeout) * time.Second
	writeTimeout := time.Duration(conf.WriteTimeout) * time.Second

	adds := strings.Split(conf.Addr, ",")
	conn := redisBase.NewUniversalClient(&redisBase.UniversalOptions{
		Addrs:         adds,
		ReadOnly:      true,
		RouteRandomly: true,
		PoolSize:      conf.MaxActive,
		Password:      conf.Password,
		ReadTimeout:   readTimeout,
		WriteTimeout:  writeTimeout,
		PoolTimeout:   connectTimeout,
		IdleTimeout:   idleTimeout,
	})
	_, err := conn.Ping(context.TODO()).Result()
	if err != nil {
		return nil, err
	}
	return conn, err
}
