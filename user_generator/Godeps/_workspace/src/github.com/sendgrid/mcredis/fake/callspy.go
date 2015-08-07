package fake

type CallSpy struct {
	CalledParams map[string]interface{}
	CallCount    int
}
