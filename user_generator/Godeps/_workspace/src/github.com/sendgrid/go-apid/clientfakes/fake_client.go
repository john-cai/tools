package clientfakes

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"sync"

	"github.com/sendgrid/go-apid"
)

type funcSpy struct {
	function     func(url.Values) (interface{}, error)
	Called       bool
	CalledParams url.Values
}

func (s *funcSpy) On(f func(params url.Values) (interface{}, error)) *funcSpy {
	s.function = func(params url.Values) (interface{}, error) {
		s.Called = true
		s.CalledParams = params
		return f(params)
	}

	return s
}

type FakeClient struct {
	functions map[apid.APIdFunction]func(url.Values) (interface{}, error)
	mutex     sync.Mutex
}

func NewFakeClient() *FakeClient {
	return &FakeClient{
		functions: make(map[apid.APIdFunction]func(url.Values) (interface{}, error)),
	}
}

func (fc *FakeClient) DoFunction(name apid.APIdFunction, params url.Values, dataPtr interface{}) error {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()
	f, ok := fc.functions[name]
	if !ok {
		return errors.New(fmt.Sprintf("Could not find APId function: %s", name))
	}
	val, err := f(params)
	if err != nil {
		return err
	}

	x := reflect.Indirect(reflect.ValueOf(dataPtr))
	if val != nil {
		x.Set(reflect.ValueOf(val))
	}
	return nil
}

func (fc *FakeClient) RegisterFunction(name apid.APIdFunction, function func(params url.Values) (interface{}, error)) *funcSpy {
	spy := new(funcSpy).On(function)
	fc.functions[name] = spy.function
	return spy
}

func (fc *FakeClient) Name() string {
	return "apid-fake"
}

func (fc *FakeClient) Check() error {
	return nil
}
