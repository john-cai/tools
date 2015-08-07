package fake

type FakeConnection struct {
	do func(commandName string, args ...interface{}) (reply interface{}, err error)
}

func (f *FakeConnection) Close() error {
	return nil
}

func (f *FakeConnection) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	if f.do != nil {
		return f.do(commandName, args)
	}

	return "success", nil
}

func (f *FakeConnection) RespondToDo(impl func(commandName string, args ...interface{}) (reply interface{}, err error)) {
	f.do = impl
}
