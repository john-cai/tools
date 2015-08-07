package mcredis

type NoMasterNameError struct {
	s string
}

func (n NoMasterNameError) Error() string {
	return n.s
}

type MasterNameNotKnownError struct {
	s string
}

func (m MasterNameNotKnownError) Error() string {
	return m.s
}

type MalformedServerNodeError struct {
	s string
}

func (m MalformedServerNodeError) Error() string {
	return m.s
}

type NoSentinelFound struct {
	s string
}

func (n NoSentinelFound) Error() string {
	return n.s
}
