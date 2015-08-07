package ln

import (
	"encoding/json"
	"errors"
	"log/syslog"
	"reflect"
	"testing"
)

type Event struct {
	Event string
	Date  string
}

type mockSysLog map[syslog.Priority]string

type MockSysLogger struct {
	log         mockSysLog
	level       syslog.Priority
	returnError bool
}

func (_ *MockSysLogger) Close() {}

func (m *MockSysLogger) Emerg(message string) error {
	m.log[m.level] = message
	if m.returnError {
		return errors.New("error when calling Emerg " + message)
	}
	return nil
}
func (m *MockSysLogger) Alert(message string) error {
	m.log[m.level] = message
	if m.returnError {
		return errors.New("error when calling Alert " + message)
	}
	return nil
}
func (m *MockSysLogger) Crit(message string) error {
	m.log[m.level] = message
	if m.returnError {
		return errors.New("error when calling Crit " + message)
	}
	return nil
}
func (m *MockSysLogger) Err(message string) error {
	m.log[m.level] = message
	if m.returnError {
		return errors.New("error when calling Err " + message)
	}
	return nil
}
func (m *MockSysLogger) Notice(message string) error {
	m.log[m.level] = message
	if m.returnError {
		return errors.New("error when calling Notice " + message)
	}
	return nil
}
func (m *MockSysLogger) Warning(message string) error {
	m.log[m.level] = message
	if m.returnError {
		return errors.New("error when calling Warning " + message)
	}
	return nil
}
func (m *MockSysLogger) Info(message string) error {
	m.log[m.level] = message
	if m.returnError {
		return errors.New("error when calling Info " + message)
	}
	return nil
}
func (m *MockSysLogger) Debug(message string) error {
	m.log[m.level] = message
	if m.returnError {
		return errors.New("error when calling Debug " + message)
	}
	return nil
}

// if lower (than specified in constructor) level is passed, output is nil
type LevelLogTest struct {
	message  string
	values   Map
	level    syslog.Priority
	facility syslog.Priority
	function func(string, Map, LevelLogger) error
}

var levelLogTests = []LevelLogTest{

	LevelLogTest{
		message:  "logging emerg",
		values:   make(Map),
		level:    syslog.LOG_EMERG,
		facility: syslog.LOG_AUTH,
		function: func(message string, values Map, l LevelLogger) error {
			return l.Emerg(message, values)
		},
	},
	LevelLogTest{
		message:  "logging alert",
		values:   make(Map),
		level:    syslog.LOG_ALERT,
		facility: syslog.LOG_AUTH,
		function: func(message string, values Map, l LevelLogger) error {
			return l.Alert(message, values)
		},
	},
	LevelLogTest{
		message:  "logging crit",
		values:   make(Map),
		level:    syslog.LOG_CRIT,
		facility: syslog.LOG_AUTH,
		function: func(message string, values Map, l LevelLogger) error {
			return l.Crit(message, values)
		},
	},
	LevelLogTest{
		message:  "logging err",
		values:   make(Map),
		level:    syslog.LOG_ERR,
		facility: syslog.LOG_AUTH,
		function: func(message string, values Map, l LevelLogger) error {
			return l.Err(message, values)
		},
	},
	LevelLogTest{
		message:  "logging warning",
		values:   make(Map),
		level:    syslog.LOG_WARNING,
		facility: syslog.LOG_AUTH,
		function: func(message string, values Map, l LevelLogger) error {
			return l.Warning(message, values)
		},
	},
	LevelLogTest{
		message:  "logging notice",
		values:   make(Map),
		level:    syslog.LOG_NOTICE,
		facility: syslog.LOG_AUTH,
		function: func(message string, values Map, l LevelLogger) error {
			return l.Notice(message, values)
		},
	},
	LevelLogTest{
		message:  "logging info",
		values:   make(Map),
		level:    syslog.LOG_INFO,
		facility: syslog.LOG_AUTH,
		function: func(message string, values Map, l LevelLogger) error {
			return l.Info(message, values)
		},
	},
	LevelLogTest{
		message:  "logging debug",
		values:   make(Map),
		level:    syslog.LOG_DEBUG,
		facility: syslog.LOG_MAIL,
		function: func(message string, values Map, l LevelLogger) error {
			return l.Debug(message, values)
		},
	},
}

// test expected log message is converted to correct JSON
func TestLevelLogging(t *testing.T) {
	// tests all the logging levels
	for _, test := range levelLogTests {
		logResults := make(mockSysLog)
		mockSysLogger := &MockSysLogger{log: logResults, level: test.level} // a Syslogger

		testLogger := newJSONLogger(mockSysLogger, test.level, test.facility)
		test.function(test.message, test.values, testLogger)

		expectedLogMessage := test.values
		expectedLogMessage["message"] = test.message

		marshalled, err := json.Marshal(expectedLogMessage)

		if err != nil {
			t.Errorf("could not marshal json %s", err.Error())
		}
		expectedLogString := string(marshalled)

		if logResults[test.level] != expectedLogString {
			t.Errorf("level:%v, expected %s, got %s", test.level, expectedLogString, logResults[test.level])
		}
	}
}

// test logging below (less critical than) provided level is not logged
func TestLevelLoggingLessCritical(t *testing.T) {
	for _, test := range levelLogTests {
		if test.level == syslog.LOG_EMERG {
			continue // nothing is more critical than LOG_EMERG
		}
		logResults := make(mockSysLog)
		mockSysLogger := &MockSysLogger{log: logResults, level: test.level} // a Syslogger
		testLogger := newJSONLogger(mockSysLogger, test.level-1, test.facility)
		test.function(test.message, test.values, testLogger)

		if !reflect.DeepEqual(logResults, make(mockSysLog)) {
			t.Error("expected empty logResults, got non-empty logResults")
		}
	}
}

// test logging above (more critical than) provided level is logged
func TestLevelLoggingMoreCritical(t *testing.T) {
	for _, test := range levelLogTests {
		if test.level == syslog.LOG_DEBUG {
			continue // nothing is less critical than LOG_DEBUG
		}
		logResults := make(mockSysLog)
		mockSysLogger := &MockSysLogger{log: logResults, level: test.level} // a Syslogger
		testLogger := newJSONLogger(mockSysLogger, test.level+1, test.facility)
		test.function(test.message, test.values, testLogger)

		expectedLogMessage := test.values
		expectedLogMessage["message"] = test.message

		marshalled, err := json.Marshal(expectedLogMessage)

		if err != nil {
			t.Errorf("could not marshal json %s", err.Error())
		}
		expectedLogString := string(marshalled)

		if logResults[test.level] != expectedLogString {
			t.Errorf("level:%v, expected %s, got %s", test.level, expectedLogString, logResults[test.level])
		}
	}
}

// test if SetLevel correctly updates the logger's logging level
func TestSetLevel(t *testing.T) {
	logResults := make(mockSysLog)
	mockSysLogger := &MockSysLogger{log: logResults, level: syslog.LOG_DEBUG}
	testLogger := newJSONLogger(mockSysLogger, syslog.LOG_DEBUG, syslog.LOG_AUTH)
	testLogger.SetLevel(syslog.LOG_EMERG)
	level := testLogger.Level()
	if level != syslog.LOG_EMERG {
		t.Errorf("testLogger.Level() not returning the right level: expected LOG_EMERG, got %v", level)
	}
}

// test if error handler gets called correctly in the event that the syslogger call has an error
func TestErrorHandler(t *testing.T) {
	logResults := make(mockSysLog)
	mockSysLogger := &MockSysLogger{log: logResults, level: syslog.LOG_DEBUG, returnError: true}
	testLogger := newJSONLogger(mockSysLogger, syslog.LOG_DEBUG, syslog.LOG_AUTH)
	errorHandlerCalled := false
	testLogger.SetErrorHandler(func(error) {
		errorHandlerCalled = true
	})
	testLogger.Alert("test alert log", nil)
	if !errorHandlerCalled {
		t.Error("errorHandler should be called when output() errors")
	}
}

// test that logging call correctly returns an error when syslogger errors
func TestErrorWhenNoHandler(t *testing.T) {
	logResults := make(mockSysLog)
	mockSysLogger := &MockSysLogger{log: logResults, level: syslog.LOG_DEBUG, returnError: true}
	testLogger := newJSONLogger(mockSysLogger, syslog.LOG_DEBUG, syslog.LOG_AUTH)

	err := testLogger.Alert("test alert log", nil)
	if err == nil {
		t.Error("Alert should have returned an error")
	}
}

// test that we can query the facility
func TestLogFacility(t *testing.T) {
	logResults := make(mockSysLog)
	mockSysLogger := &MockSysLogger{log: logResults, level: syslog.LOG_DEBUG, returnError: true}
	testLogger := newJSONLogger(mockSysLogger, syslog.LOG_DEBUG, syslog.LOG_AUTH)

	facility := testLogger.Facility()
	if facility != syslog.LOG_AUTH {
		t.Error("Facility priority does not match")
	}
}

// test that we can query the facility
func TestLevelName(t *testing.T) {
	logResults := make(mockSysLog)
	mockSysLogger := &MockSysLogger{log: logResults, level: syslog.LOG_DEBUG, returnError: true}
	testLogger := newJSONLogger(mockSysLogger, syslog.LOG_DEBUG, syslog.LOG_AUTH)

	err := testLogger.SetLevelName("BOGUS")
	if err == nil {
		t.Error("Expected an error for setting an invalid level name")
	}

	err = testLogger.SetLevelName("DEBUG")
	if err != nil {
		t.Error("Expected no error for setting a level name")
	}
}

func TestEvent(t *testing.T) {
	logResults := make(mockSysLog)
	mockSysLogger := &MockSysLogger{log: logResults, level: syslog.LOG_INFO} // a Syslogger

	testLogger := newJSONLogger(mockSysLogger, syslog.LOG_INFO, syslog.LOG_AUTH)
	logValues := Map{"k1": "v1"}
	testLogger.Event("testEvent", logValues)

	expectedLogMessage := logValues
	marshalled, err := json.Marshal(expectedLogMessage)
	if err != nil {
		t.Errorf("could not marshal json %s", err.Error())
	}

	expectedLogString := string(marshalled)
	if logResults[EVENT_LOG_LEVEL] != expectedLogString {
		t.Errorf("expected %s, got %s", expectedLogString, logResults[EVENT_LOG_LEVEL])
	}
}
