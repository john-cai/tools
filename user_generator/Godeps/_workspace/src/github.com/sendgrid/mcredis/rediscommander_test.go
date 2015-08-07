package mcredis_test

import (
	"errors"

	"github.com/garyburd/redigo/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sendgrid/mcredis"
	"github.com/sendgrid/mcredis/fake"
)

var _ = Describe("RedisCommander", func() {
	var (
		commander      *mcredis.RedisInstance
		fakeConnection *fake.FakeConnection
	)

	BeforeEach(func() {
		fakeConnection = &fake.FakeConnection{}
		commander = &mcredis.RedisInstance{fake.NewFakeCommander()}
	})

	Describe("#Set", func() {
		It("calls SET on the redis connection", func() {
			var setCalled bool
			fakeConnection.RespondToDo(func(command string, args ...interface{}) (interface{}, error) {
				setCalled = true
				Expect(command).To(Equal("SET"))
				Expect(args[0]).To(HaveLen(2))
				Expect(args[0].([]interface{})[0]).To(Equal("my key"))
				Expect(args[0].([]interface{})[1]).To(Equal("my value"))

				return nil, nil
			})

			err := commander.Set("my key", "my value")
			Expect(err).To(BeNil())
			Expect(setCalled).To(BeTrue())
		})

		It("returns the redis error", func() {
			fakeConnection.RespondToDo(func(command string, args ...interface{}) (interface{}, error) {
				return nil, errors.New("set error")
			})

			err := commander.Set("my key", "my value")
			Expect(err.Error()).To(Equal("set error"))
		})
	})

	Describe("#Del", func() {
		It("calls DEL on the redis connection", func() {
			var delCalled bool
			fakeConnection.RespondToDo(func(command string, args ...interface{}) (interface{}, error) {
				delCalled = true
				Expect(command).To(Equal("DEL"))
				Expect(args[0]).To(HaveLen(1))
				Expect(args[0].([]interface{})[0]).To(Equal("my key"))

				return nil, nil
			})

			err := commander.Del("my key")
			Expect(err).To(BeNil())
			Expect(delCalled).To(BeTrue())
		})

		It("returns the redis error", func() {
			fakeConnection.RespondToDo(func(command string, args ...interface{}) (interface{}, error) {
				return nil, errors.New("del error")
			})

			err := commander.Del("my key")
			Expect(err.Error()).To(Equal("del error"))
		})
	})

	Describe("#Decr", func() {
		It("calls DECR on the redis connection", func() {
			var decrCalled bool
			fakeConnection.RespondToDo(func(command string, args ...interface{}) (interface{}, error) {
				decrCalled = true
				Expect(command).To(Equal("DECR"))
				Expect(args[0]).To(HaveLen(1))
				Expect(args[0].([]interface{})[0]).To(Equal("my key"))

				return int64(42), nil
			})

			val, err := commander.Decr("my key")
			Expect(err).To(BeNil())
			Expect(decrCalled).To(BeTrue())
			Expect(val).To(Equal(42))
		})

		It("returns the redis error", func() {
			fakeConnection.RespondToDo(func(command string, args ...interface{}) (interface{}, error) {
				return nil, errors.New("decr error")
			})

			_, err := commander.Decr("my key")
			Expect(err.Error()).To(Equal("decr error"))
		})
	})

	Describe("#Incr", func() {
		It("calls INCR on the redis connection", func() {
			var incrCalled bool
			fakeConnection.RespondToDo(func(command string, args ...interface{}) (interface{}, error) {
				incrCalled = true
				Expect(command).To(Equal("INCR"))
				Expect(args[0]).To(HaveLen(1))
				Expect(args[0].([]interface{})[0]).To(Equal("my key"))

				return int64(42), nil
			})

			val, err := commander.Incr("my key")
			Expect(err).To(BeNil())
			Expect(incrCalled).To(BeTrue())
			Expect(val).To(Equal(42))
		})

		It("returns the redis error", func() {
			fakeConnection.RespondToDo(func(string, ...interface{}) (interface{}, error) {
				return nil, errors.New("incr error")
			})

			_, err := commander.Incr("my key")
			Expect(err.Error()).To(Equal("incr error"))
		})
	})

	Describe("#Get", func() {
		It("calls GET on the redis connection", func() {
			var getCalled bool
			fakeConnection.RespondToDo(func(command string, args ...interface{}) (interface{}, error) {
				getCalled = true
				Expect(command).To(Equal("GET"))
				Expect(args[0]).To(HaveLen(1))
				Expect(args[0].([]interface{})[0]).To(Equal("my key"))

				return int64(42), nil
			})

			val, err := commander.Get("my key")
			Expect(err).To(BeNil())
			Expect(getCalled).To(BeTrue())

			myNum, err := redis.Int(val, err)
			Expect(myNum).To(Equal(42))
			Expect(err).To(BeNil())
		})
	})

	Describe("#GetInt", func() {
		It("calls GET on the redis connection and returns an int", func() {
			var getIntCalled bool
			fakeConnection.RespondToDo(func(command string, args ...interface{}) (interface{}, error) {
				getIntCalled = true
				Expect(command).To(Equal("GET"))
				Expect(args[0]).To(HaveLen(1))
				Expect(args[0].([]interface{})[0]).To(Equal("my key"))

				return int64(42), nil
			})

			val, err := commander.GetInt("my key")
			Expect(err).To(BeNil())
			Expect(getIntCalled).To(BeTrue())
			Expect(val).To(Equal(42))
		})
	})

	Describe("#Expire", func() {
		It("calls EXPIRE on the redis connection", func() {
			var getCalled bool
			fakeConnection.RespondToDo(func(command string, args ...interface{}) (interface{}, error) {
				getCalled = true
				Expect(command).To(Equal("EXPIRE"))
				Expect(args[0]).To(HaveLen(2))
				Expect(args[0].([]interface{})[0]).To(Equal("my key"))
				Expect(args[0].([]interface{})[1]).To(Equal(100))

				return int64(1), nil
			})
			val, err := commander.Expire("my key", 100)
			Expect(err).To(BeNil())
			Expect(getCalled).To(BeTrue())
			Expect(val).To(Equal(1))
		})
	})

	Describe("#Exists", func() {
		It("calls EXISTS on the redis connection and returns a bool", func() {
			var existsCalled bool
			fakeConnection.RespondToDo(func(command string, args ...interface{}) (interface{}, error) {
				existsCalled = true
				Expect(command).To(Equal("EXISTS"))
				Expect(args[0]).To(HaveLen(1))
				Expect(args[0].([]interface{})[0]).To(Equal("my key"))

				return int64(1), nil
			})
			val, err := commander.Exists("my key")
			Expect(err).To(BeNil())
			Expect(existsCalled).To(BeTrue())
			Expect(val).To(BeTrue())
		})
	})

	Describe("#SetAdd", func() {
		It("calls SADD on the redis connection", func() {
			saddCalled := false
			elements := []string{"one", "two"}

			fakeConnection.RespondToDo(func(command string, args ...interface{}) (interface{}, error) {
				saddCalled = true
				Expect(command).To(Equal("SADD"))
				Expect(args[0]).To(HaveLen(2))
				Expect(args[0].([]interface{})[0]).To(Equal("my key"))
				Expect(args[0].([]interface{})[1]).To(Equal(elements))

				return int64(1), nil
			})

			err := commander.SetAdd("my key", elements...)

			Expect(err).To(BeNil())
			Expect(saddCalled).To(BeTrue())
		})
	})

	Describe("#GetMembersOfSet", func() {
		It("calls SMEMBERS on the redis connection and returns a list of strings", func() {
			var getMembersCalled bool
			fakeConnection.RespondToDo(func(command string, args ...interface{}) (interface{}, error) {
				getMembersCalled = true
				Expect(command).To(Equal("SMEMBERS"))
				Expect(args[0]).To(HaveLen(1))
				Expect(args[0].([]interface{})[0]).To(Equal("my key"))

				return []interface{}{[]byte("[one two]")}, nil
			})

			val, err := commander.GetMembersOfSet("my key")
			Expect(err).To(BeNil())
			Expect(getMembersCalled).To(BeTrue())
			Expect(val).To(Equal([]string{"one", "two"}))
		})
	})
})
