package mcredis_test

import (
	"errors"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sendgrid/mcredis"
	"github.com/sendgrid/mcredis/fake"
)

var _ = Describe("RedisSentinel", func() {

	var (
		dialSentinelCount int
		dialMasterCount   int
		redisDialer       *fake.FakeDialer
		redis             *fake.FakeConnection
		sentinel          *mcredis.RedisInstance
		roleCalled        bool
	)

	BeforeEach(func() {
		redisDialer = fake.NewFakeDialer()
		redis = &fake.FakeConnection{}
		dialSentinelCount = 0
		dialMasterCount = 0
		roleCalled = false

		redisDialer.RespondToDial(func(url, address string) (*fake.FakeConnection, error) {
			dialMasterCount++
			if address == "badMasterAddr:port" {
				return nil, errors.New("master connection error")
			}
			return redis, nil
		})

		redisDialer.RespondToDialTimeout(func(url, address string, connectTimeout, readTimeout, writeTimeout time.Duration) (*fake.FakeConnection, error) {
			dialSentinelCount++
			if address == "badSentinel" {
				return nil, errors.New("sentinel connection error")
			}
			return redis, nil
		})

		redis.RespondToDo(func(commandName string, args ...interface{}) (reply interface{}, err error) {
			if commandName == "SENTINEL" {
				return []interface{}{"masterAddr", "port"}, nil
			} else {
				roleCalled = true
				return "role:master", nil
			}
		})
	})

	Describe("#ConnectAndGetMaster", func() {
		It("connects to a sentinel and returns the master connection", func() {
			sentinel = mcredis.NewSentinel(redisDialer, "master", "goodSentinel", 0)
			masterConn, err := sentinel.ConnectAndGetMaster()

			Expect(err).To(BeNil())
			Expect(dialSentinelCount).To(Equal(1))
			Expect(dialMasterCount).To(Equal(1))
			Expect(roleCalled).To(BeTrue())
			Expect(masterConn).To(Equal(redis))
		})

		It("returns an error if it cannot connect to any sentinels", func() {
			sentinel = mcredis.NewSentinel(redisDialer, "master", "badSentinel", 0)
			_, err := sentinel.ConnectAndGetMaster()

			Expect(err.Error()).To(Equal("Failed to connect to a sentinel"))
			Expect(dialSentinelCount).To(Equal(1))
		})

		It("tries to connect to the next sentinel if connected master is not master", func() {
			redis.RespondToDo(func(commandName string, args ...interface{}) (reply interface{}, err error) {
				if commandName == "SENTINEL" {
					return []interface{}{"masterAddr", "port"}, nil
				} else {
					roleCalled = true
					return "role:slave", nil
				}
			})
			sentinel = mcredis.NewSentinel(redisDialer, "master", "goodSentinel,goodSentinel", 0)
			masterConn, err := sentinel.ConnectAndGetMaster()

			Expect(dialSentinelCount).To(Equal(2))
			Expect(dialMasterCount).To(Equal(2))
			Expect(err.Error()).To(Equal("master role confirmation error"))
			Expect(masterConn).To(BeNil())
		})

		It("tries to connect to the next sentinel if connecting to master, after 3 failed attempts and waiting the redis-configured amount of time, returns an error", func() {
			redis.RespondToDo(func(commandName string, args ...interface{}) (reply interface{}, err error) {
				return []interface{}{"badMasterAddr", "port"}, nil
			})
			sentinel = mcredis.NewSentinel(redisDialer, "master", "goodSentinel,goodSentinel", 0)
			masterConn, err := sentinel.ConnectAndGetMaster()

			Expect(dialSentinelCount).To(Equal(2))
			Expect(dialMasterCount).To(Equal(6))
			Expect(err.Error()).To(Equal("master connection error"))
			Expect(masterConn).To(BeNil())
		})

		It("does not fail if 2nd call to connect to master succeeds", func() {
			redisDialer.RespondToDial(func(connType, url string) (mcredis.Connection, error) {
				dialMasterCount++
				if dialMasterCount == 2 {
					return redis, nil
				} else {
					return redis, errors.New("failed to connect to redis")
				}
			})
			sentinel = mcredis.NewSentinel(redisDialer, "master", "goodSentinel,goodSentinel", 0)
			masterConn, err := sentinel.ConnectAndGetMaster()

			Expect(dialSentinelCount).To(Equal(1))
			Expect(dialMasterCount).To(Equal(2))
			Expect(err).To(BeNil())
			Expect(masterConn).ToNot(BeNil())
		})

		It("tries to connect to the next sentinel if the connected sentinel returns a nil master address", func() {
			redis.RespondToDo(func(commandName string, args ...interface{}) (reply interface{}, err error) {
				return nil, nil
			})
			sentinel = mcredis.NewSentinel(redisDialer, "unknownMaster", "goodSentinel,goodSentinel", 0)
			masterConn, err := sentinel.ConnectAndGetMaster()

			Expect(dialSentinelCount).To(Equal(2))
			Expect(masterConn).To(BeNil())
			Expect(err.Error()).To(Equal("Sentinel does not know the master: unknownMaster"))
		})

		It("tries to connect to the next sentinel if getting master address returns an error", func() {
			getMasterAddrCallCount := 0
			redis.RespondToDo(func(commandName string, args ...interface{}) (reply interface{}, err error) {
				getMasterAddrCallCount++
				return []interface{}{"something", "port"}, errors.New("error getting master address")
			})
			sentinel = mcredis.NewSentinel(redisDialer, "master", "goodSentinel,goodSentinel", 0)
			_, err := sentinel.ConnectAndGetMaster()

			Expect(dialSentinelCount).To(Equal(2))
			Expect(getMasterAddrCallCount).To(Equal(2))
			Expect(err.Error()).To(Equal("error getting master address"))
		})

		It("puts successfully-connected sentinel address at beginning of sentinel servers list", func() {
			sentinel = mcredis.NewSentinel(redisDialer, "master", "badSentinel,badSentinel,goodSentinel,badSentinel", 0)
			_, err := sentinel.ConnectAndGetMaster()

			Expect(err).To(BeNil())
			Expect(sentinel.GetServers()).To(Equal([]string{"goodSentinel", "badSentinel", "badSentinel", "badSentinel"}))
			Expect(dialSentinelCount).To(Equal(3))
		})
	})

	Describe("#GetInfo", func() {
		It("connects to a sentinel and returns the master address and sentinel address that it used", func() {
			sentinel = mcredis.NewSentinel(redisDialer, "master", "goodSentinel", 0)
			master, sent := sentinel.GetInfo()

			Expect(master).To(Equal("masterAddr:port"))
			Expect(sent).To(Equal("goodSentinel"))
		})
	})
})
