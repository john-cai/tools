package mcredis

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

type RedisCommander interface {
	Do(string, ...interface{}) (interface{}, error)
	GetInfo() (string, string)
}

type RedisConfig struct {
	MasterName  string
	Nodes       []string // Must be a list of dial-able hosts containing at least one master
	CallTimeout time.Duration
	Timeout     time.Duration
	PoolSize    int
	MaxIdle     int
}

type RedisInstance struct {
	RedisCommander
}

func NewMcRedis(conf RedisConfig) (RedisInstance, error) {
	rs, err := NewSentinel(conf)
	if err != nil {
		return RedisInstance{}, err
	}
	return RedisInstance{rs}, nil
}

func (r *RedisInstance) Set(key string, value interface{}) error {
	_, err := r.Do("SET", key, value)
	return err
}

func (r *RedisInstance) Del(key string) error {
	_, err := r.Do("DEL", key)
	return err
}

func (r *RedisInstance) Decr(key string) (int, error) {
	return redis.Int(r.Do("DECR", key))
}

func (r *RedisInstance) Incr(key string) (int, error) {
	return redis.Int(r.Do("INCR", key))
}

func (r *RedisInstance) Get(key string) (interface{}, error) {
	return r.Do("GET", key)
}

func (r *RedisInstance) GetInt(key string) (int, error) {
	return redis.Int(r.Do("GET", key))
}

func (r *RedisInstance) Expire(key string, ttl int) (int, error) {
	return redis.Int(r.Do("EXPIRE", key, ttl))
}

func (r *RedisInstance) Exists(key string) (bool, error) {
	return redis.Bool(r.Do("EXISTS", key))
}

func (r *RedisInstance) SetDiff(key string, keys []string) ([]string, error) {
	a := make([]interface{}, len(keys)+1)
	a[0] = key
	for i, k := range keys {
		a[i+1] = k
	}
	return redis.Strings(r.Do("SDIFF", a...))
}

func (r *RedisInstance) GetMembersOfSet(key string) ([]string, error) {
	return redis.Strings(r.Do("SMEMBERS", key))
}
