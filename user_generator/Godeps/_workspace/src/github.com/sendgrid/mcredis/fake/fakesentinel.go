package fake

type FakeSentinel struct {
	calls               map[string]CallSpy
	connectAndGetMaster func() (FakeConnection, error)
}

func NewFakeSentinel() *FakeSentinel {
	return &FakeSentinel{
		calls: map[string]CallSpy{},
	}
}

func (f *FakeSentinel) CallTo(funcName string) CallSpy {
	return f.calls[funcName]
}

func (f *FakeSentinel) ConnectAndGetMaster() (FakeConnection, error) {
	callCount := f.calls["ConnectAndGetMaster"].CallCount
	callCount++
	f.calls["ConnectAndGetMaster"] = CallSpy{
		CallCount: callCount,
	}

	if f.connectAndGetMaster != nil {
		return f.connectAndGetMaster()
	}
	return FakeConnection{}, nil
}

func (f *FakeSentinel) RespondToConnectAndGetMaster(impl func() (FakeConnection, error)) {
	f.connectAndGetMaster = impl
}
