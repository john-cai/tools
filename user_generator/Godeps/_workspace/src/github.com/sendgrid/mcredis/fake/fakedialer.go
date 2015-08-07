package fake

import "time"

type FakeDialer struct {
	Calls       map[string]CallSpy
	dial        func(string, string) (*FakeConnection, error)
	dialTimeout func(string, string, time.Duration, time.Duration, time.Duration) (*FakeConnection, error)
}

func NewFakeDialer() *FakeDialer {
	return &FakeDialer{
		Calls: map[string]CallSpy{},
	}
}

func (f *FakeDialer) Dial(connectionType string, address string) (*FakeConnection, error) {
	callCount := f.Calls["Dial"].CallCount
	callCount++

	f.Calls["Dial"] = CallSpy{
		CallCount: callCount,
	}

	return f.dial(connectionType, address)
}

func (f *FakeDialer) RespondToDial(impl func(string, string) (*FakeConnection, error)) {
	f.dial = impl
}

func (f *FakeDialer) DialTimeout(
	connectionType string,
	address string,
	connectTimeout,
	readTimeout,
	writeTimeout time.Duration,
) (*FakeConnection, error) {

	callCount := f.Calls["DialTimeout"].CallCount
	callCount++

	f.Calls["DialTimeout"] = CallSpy{
		CallCount: callCount,
	}

	return f.dialTimeout(connectionType, address, connectTimeout, readTimeout, writeTimeout)
}

func (f *FakeDialer) RespondToDialTimeout(impl func(string, string, time.Duration, time.Duration, time.Duration) (*FakeConnection, error)) {
	f.dialTimeout = impl
}
