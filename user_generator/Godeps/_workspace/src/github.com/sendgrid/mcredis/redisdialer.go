package mcredis

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

type RedisDialer interface {
	Dial(string, string) (Connection, error)
	DialTimeout(string, string, time.Duration, time.Duration, time.Duration) (Connection, error)
}

type McRedisDialer struct{}

func (r McRedisDialer) Dial(connectionType, url string) (Connection, error) {
	return redis.Dial(connectionType, url)
}

func (r McRedisDialer) DialTimeout(connectionType string, address string, connectTimeout, readTimeout, writeTimeout time.Duration) (Connection, error) {
	return redis.DialTimeout(connectionType, address, connectTimeout, readTimeout, writeTimeout)
}
