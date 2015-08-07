package ln

import (
	"errors"
	"log/syslog"
	"reflect"
	"sync"
)

type logTracking map[syslog.Priority][]message

// logger that records all message and enables asserting
// for testing purposes
type MockLogger struct {
	tracking logTracking
	level    syslog.Priority
	facility syslog.Priority
	callback func(error)
	Exited   bool

	sync.RWMutex
}

func NewMockLogger() *MockLogger {
	return &MockLogger{
		tracking: make(logTracking),
	}
}

func (m *MockLogger) Emerg(msg string, v Map) error {
	m.store(syslog.LOG_EMERG, msg, v)
	return nil
}

func (m *MockLogger) Alert(msg string, v Map) error {
	m.store(syslog.LOG_ALERT, msg, v)
	return nil
}

func (m *MockLogger) Crit(msg string, v Map) error {
	m.store(syslog.LOG_CRIT, msg, v)
	return nil
}

func (m *MockLogger) Err(msg string, v Map) error {
	m.store(syslog.LOG_ERR, msg, v)
	return nil
}

func (m *MockLogger) Notice(msg string, v Map) error {
	m.store(syslog.LOG_NOTICE, msg, v)
	return nil
}

func (m *MockLogger) Warning(msg string, v Map) error {
	m.store(syslog.LOG_WARNING, msg, v)
	return nil
}

func (m *MockLogger) Info(msg string, v Map) error {
	m.store(syslog.LOG_INFO, msg, v)
	return nil
}

func (m *MockLogger) Debug(msg string, v Map) error {
	m.store(syslog.LOG_DEBUG, msg, v)
	return nil
}

func (m *MockLogger) Fatal(msg string, v Map) {
	m.Exited = true
	m.store(syslog.LOG_EMERG, msg, v)
}

func (m *MockLogger) Event(event string, v Map) error {
	return nil
}

func (_ *MockLogger) Close() {}

// we will need to make this thread-safe in case logger is called asychronously
func (m *MockLogger) store(p syslog.Priority, msg string, v Map) {
	m.Lock()
	if m.tracking[p] == nil {
		m.tracking[p] = make([]message, 0, 1)
	}
	m.tracking[p] = append(m.tracking[p], message{Message: msg, Values: v})
	m.Unlock()
}

func (m *MockLogger) SetErrorHandler(callback func(error)) {
	m.callback = callback
}

func (m *MockLogger) SetLevel(l syslog.Priority) error {
	if l < 0 || l > syslog.LOG_DEBUG {
		return errors.New("unrecognized level " + string(l))
	}
	m.level = l
	return nil
}

func (m *MockLogger) SetLevelName(l string) error {
	level, err := stringToLevel(l)

	if err != nil {
		return err
	}

	m.level = level
	return nil
}

func (m *MockLogger) Level() syslog.Priority {
	return m.level

}
func (m *MockLogger) Facility() syslog.Priority {
	return m.facility

}

type TestingT interface {
	Errorf(format string, args ...interface{})
	Logf(format string, args ...interface{})
}

// Utility function to verify that a value was encoded into the logger
func (m *MockLogger) AssertLogged(t TestingT, p syslog.Priority, msg string, values Map) {
	// look through the tracked log messages and find if there is an
	// artifact that matches

	for _, s := range m.tracking[p] {
		equal := reflect.DeepEqual(message{Message: msg, Values: values}, s)
		if equal {
			return
		}
	}

	t.Errorf("No matches found for %+v\n", values)
}

func (m *MockLogger) AssertMessageLogged(t TestingT, p syslog.Priority, msg string) {

	for _, s := range m.tracking[p] {
		equal := s.Message == msg

		if equal {
			return
		}
	}

	t.Errorf("No matches found for %s\n", msg)

}

// Utility function to print all values from a priority
func (m *MockLogger) Dump(t TestingT, p syslog.Priority) {
	for _, s := range m.tracking[p] {
		t.Logf("tracking is %+v", s)
	}
}

type nullLogger struct{}

func (n *nullLogger) Close()                               {}
func (n *nullLogger) Emerg(message string, v Map) error    { return nil }
func (n *nullLogger) Alert(message string, v Map) error    { return nil }
func (n *nullLogger) Crit(message string, v Map) error     { return nil }
func (n *nullLogger) Err(message string, v Map) error      { return nil }
func (n *nullLogger) Notice(message string, v Map) error   { return nil }
func (n *nullLogger) Warning(message string, v Map) error  { return nil }
func (n *nullLogger) Info(message string, v Map) error     { return nil }
func (n *nullLogger) Debug(message string, v Map) error    { return nil }
func (n *nullLogger) Fatal(message string, v Map)          {}
func (n *nullLogger) SetErrorHandler(callback func(error)) {}
func (n *nullLogger) SetLevel(l syslog.Priority) error     { return nil }
func (n *nullLogger) SetLevelName(l string) error          { return nil }
func (n *nullLogger) Level() syslog.Priority               { return syslog.LOG_INFO }
func (n *nullLogger) Facility() syslog.Priority            { return syslog.LOG_INFO }

func NewNullLogger() LevelLogger {
	return &nullLogger{}
}
