package redisadaptor

import (
	"strings"
	"time"

	"github.com/pborman/uuid"
	"github.com/sendgrid/mcredis"
)

type KeyStore interface {
	SetWithUUID(data []byte, ttl int) (string, error)
	Set(key string, data []byte, ttl int) error
	Get(key string) ([]byte, error)
	Delete(key string) error
	Check() error
	Name() string
}

type redisAdaptor struct {
	config mcredis.RedisConfig
}

func New(addrs string, masterName string, timeout time.Duration, callTimeout time.Duration, poolSize int, maxIdleConnections int) *redisAdaptor {

	config := mcredis.RedisConfig{
		Nodes:       strings.Split(addrs, ","),
		MasterName:  masterName,
		Timeout:     timeout,
		CallTimeout: callTimeout,
		PoolSize:    poolSize,
		MaxIdle:     maxIdleConnections,
	}

	return &redisAdaptor{config: config}
}

func (adaptor *redisAdaptor) Name() string {
	return "redis"
}

func (adaptor *redisAdaptor) Check() error {
	// Get connection
	r, err := mcredis.NewMcRedis(adaptor.config)
	if err != nil {
		return err
	}

	// Perform health check
	_, err = r.Do("PING")
	return err
}

func (adaptor *redisAdaptor) Get(key string) ([]byte, error) {
	emptyData := make([]byte, 0)

	// Get connection
	r, err := mcredis.NewMcRedis(adaptor.config)
	if err != nil {
		return emptyData, err
	}

	dataResponse, err := r.Get(key)
	if err != nil {
		return emptyData, err
	}

	switch dataResponse.(type) {
	case nil:
		return emptyData, nil
	}

	return dataResponse.([]byte), nil
}

func (adaptor *redisAdaptor) Delete(key string) error {
	// Get connection
	r, err := mcredis.NewMcRedis(adaptor.config)
	if err != nil {
		return err
	}

	// Delete
	err = r.Del(key)
	if err != nil {
		return err
	}

	return nil
}

func (adaptor *redisAdaptor) SetWithUUID(data []byte, ttl int) (string, error) {
	// though collisions are so rare as to be non-existant, let's check anyway
	UUID, err := adaptor.availableUUID()
	if err != nil {
		return "", err
	}

	err = adaptor.Set(UUID, data, ttl)

	return UUID, err
}

func (adaptor *redisAdaptor) Set(key string, data []byte, ttl int) error {
	// Get connection
	r, err := mcredis.NewMcRedis(adaptor.config)
	if err != nil {
		return err
	}

	if ttl > 0 {
		_, err = r.Do("SETEX", key, ttl, data)
	} else {
		_, err = r.Do("SET", key, data)
	}

	if err != nil {
		return err
	}

	return nil
}

func (adaptor *redisAdaptor) availableUUID() (string, error) {
	var UUID string
	for {
		UUID = uuid.New()
		data, err := adaptor.Get(UUID)
		if err != nil {
			return "", err
		}
		// no data means key is available
		if len(data) == 0 {
			break
		}
	}
	return UUID, nil
}
