package fake

import (
	"strings"

	"github.com/sendgrid/mcredis"
	"github.com/streamrail/concurrent-map"
)

type FakeCommander struct {
	calls           cmap.ConcurrentMap
	set             func(string, interface{}) error
	getSet          func(string, interface{}) (interface{}, error)
	del             func(string) error
	get             func(string) (interface{}, error)
	getInt          func(string) (int, error)
	decr            func(string) (int, error)
	incr            func(string) (int, error)
	expire          func(string, int) (int, error)
	exists          func(string) (bool, error)
	getMembersOfSet func(string) ([]string, error)
}

func NewFakeCommander() mcredis.RedisCommander {
	return FakeCommander{
		calls: cmap.New(),
	}
}

func (f FakeCommander) GetInfo() (string, string) {
	return "", ""
}

// I'm so sorry
func (f FakeCommander) Do(action string, params ...interface{}) (interface{}, error) {
	switch strings.ToLower(action) {
	case "set":
		err := f.Set(params[0].(string), params[1])
		return nil, err
	case "get":
		return f.Get(params[0].(string))
	case "del":
		err := f.Del(params[0].(string))
		return nil, err
	case "decr":
		return f.Decr(params[0].(string))
	case "incr":
		return f.Incr(params[0].(string))
	case "getint":
		return f.GetInt(params[0].(string))
	case "expire":
		return f.Expire(params[0].(string), params[1].(int))
	case "exists":
		return f.Exists(params[0].(string))
	case "getmembersofset":
		return f.GetMembersOfSet(params[0].(string))
	case "getset":
		return f.GetSet(params[0].(string), params[1])
	default:
		return nil, nil
	}
}

func (f *FakeCommander) CallTo(funcName string) CallSpy {
	tmp, _ := f.calls.Get(funcName)

	return tmp.(CallSpy)
}

func (f *FakeCommander) Set(key string, value interface{}) error {
	call, ok := f.calls.Get("Set")
	callCount := 1
	if ok == true {
		callCount = call.(CallSpy).CallCount
		callCount++
	}
	f.calls.Set("Set", CallSpy{
		CalledParams: map[string]interface{}{
			"key":   key,
			"value": value,
		},
		CallCount: callCount,
	})

	if f.set != nil {
		return f.set(key, value)
	}
	return nil
}

func (f *FakeCommander) GetSet(key string, value interface{}) (interface{}, error) {
	call, ok := f.calls.Get("GetSet")
	callCount := 1
	if ok == true {
		callCount = call.(CallSpy).CallCount
		callCount++
	}
	f.calls.Set("GetSet", CallSpy{
		CalledParams: map[string]interface{}{
			"key":   key,
			"value": value,
		},
		CallCount: callCount,
	})

	if f.getSet != nil {
		return f.getSet(key, value)
	}
	return nil, nil
}

func (f *FakeCommander) Del(key string) error {
	call, ok := f.calls.Get("Del")
	callCount := 1
	if ok == true {
		callCount = call.(CallSpy).CallCount
		callCount++
	}
	f.calls.Set("Del", CallSpy{
		CalledParams: map[string]interface{}{
			"key": key,
		},
		CallCount: callCount,
	})

	if f.del != nil {
		return f.del(key)
	}
	return nil
}

func (f *FakeCommander) Decr(key string) (int, error) {
	call, ok := f.calls.Get("Decr")
	callCount := 1
	if ok == true {
		callCount = call.(CallSpy).CallCount
		callCount++
	}
	f.Set("Decr", CallSpy{
		CalledParams: map[string]interface{}{
			"key": key,
		},
		CallCount: callCount,
	})

	if f.decr != nil {
		return f.decr(key)
	}
	return 0, nil
}

func (f *FakeCommander) Incr(key string) (int, error) {
	call, ok := f.calls.Get("Incr")
	callCount := 1
	if ok == true {
		callCount = call.(CallSpy).CallCount
		callCount++
	}
	f.calls.Set("Incr", CallSpy{
		CalledParams: map[string]interface{}{
			"key": key,
		},
		CallCount: callCount,
	})

	if f.incr != nil {
		return f.incr(key)
	}
	return 0, nil
}

func (f *FakeCommander) Get(key string) (interface{}, error) {
	call, ok := f.calls.Get("Get")
	callCount := 1
	if ok == true {
		callCount = call.(CallSpy).CallCount
		callCount++
	}
	f.calls.Set("Get", CallSpy{
		CalledParams: map[string]interface{}{
			"key": key,
		},
		CallCount: callCount,
	})

	if f.get != nil {
		return f.get(key)
	}
	return nil, nil
}

func (f *FakeCommander) GetInt(key string) (int, error) {
	call, ok := f.calls.Get("GetInt")
	callCount := 1
	if ok == true {
		callCount = call.(CallSpy).CallCount
		callCount++
	}
	f.calls.Set("GetInt", CallSpy{
		CalledParams: map[string]interface{}{
			"key": key,
		},
		CallCount: callCount,
	})

	if f.getInt != nil {
		return f.getInt(key)
	}
	return 0, nil
}

func (f *FakeCommander) Expire(key string, ttl int) (int, error) {
	call, ok := f.calls.Get("Expire")
	callCount := 1
	if ok == true {
		callCount = call.(CallSpy).CallCount
		callCount++
	}
	f.calls.Set("Expire", CallSpy{
		CalledParams: map[string]interface{}{
			"key": key,
			"ttl": ttl,
		},
		CallCount: callCount,
	})

	if f.expire != nil {
		return f.expire(key, ttl)
	}
	return 0, nil
}

func (f *FakeCommander) Exists(key string) (bool, error) {
	call, ok := f.calls.Get("Exists")
	callCount := 1
	if ok == true {
		callCount = call.(CallSpy).CallCount
		callCount++
	}
	f.calls.Set("Exists", CallSpy{
		CalledParams: map[string]interface{}{
			"key": key,
		},
		CallCount: callCount,
	})

	if f.exists != nil {
		return f.exists(key)
	}
	return false, nil
}

func (f *FakeCommander) GetMembersOfSet(key string) ([]string, error) {
	call, ok := f.calls.Get("GetMembersOfSet")
	callCount := 1
	if ok == true {
		callCount = call.(CallSpy).CallCount
		callCount++
	}
	f.Set("GetMembersOfSet", CallSpy{
		CalledParams: map[string]interface{}{
			"key": key,
		},
		CallCount: callCount,
	})

	if f.getMembersOfSet != nil {
		return f.getMembersOfSet(key)
	}
	return nil, nil
}

func (f *FakeCommander) RespondToSet(impl func(string, interface{}) error) {
	f.set = impl
}

func (f *FakeCommander) RespondToGetSet(impl func(string, interface{}) (interface{}, error)) {
	f.getSet = impl
}

func (f *FakeCommander) RespondToDel(impl func(string) error) {
	f.del = impl
}

func (f *FakeCommander) RespondToDecr(impl func(string) (int, error)) {
	f.decr = impl
}

func (f *FakeCommander) RespondToIncr(impl func(string) (int, error)) {
	f.incr = impl
}

func (f *FakeCommander) RespondToExpire(impl func(key string, ttl int) (int, error)) {
	f.expire = impl
}

func (f *FakeCommander) RespondToGet(impl func(key string) (interface{}, error)) {
	f.get = impl
}

func (f *FakeCommander) RespondToGetInt(impl func(string) (int, error)) {
	f.getInt = impl
}

func (f *FakeCommander) RespondToExists(impl func(string) (bool, error)) {
	f.exists = impl
}

func (f *FakeCommander) RespondToGetMembersOfSet(impl func(string) ([]string, error)) {
	f.getMembersOfSet = impl
}
