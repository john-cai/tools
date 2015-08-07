package ln

import (
	"log/syslog"
	"testing"
)

func TestAssertLogged(t *testing.T) {
	mockLogger := NewMockLogger()

	err := mockLogger.Alert("this is a message", Map{"simple": "alert test", "a": "b"})
	if err != nil {
		t.Errorf("error calling 'Alert': %q", err)
	}

	mockLogger.AssertLogged(
		t,
		syslog.LOG_ALERT,
		"this is a message",
		Map{"simple": "alert test", "a": "b"})

}

func TestFatalExits(t *testing.T) {
	mockLogger := NewMockLogger()

	mockLogger.Fatal("this is a message", nil)

	if !mockLogger.Exited {
		t.Error("a call to Fatal should have exited")
	}
}

func TestMockLogLevels(t *testing.T) {
	mockLogger := NewMockLogger()

	mockLogger.Fatal("this is a message", nil)
	mockLogger.Emerg("foo", nil) // commented out so it doesn't spam/error out on chef runs, but it passes, really!
	mockLogger.Alert("foo", nil)
	mockLogger.Crit("foo", nil)
	mockLogger.Warning("foo", nil)
	mockLogger.Info("foo", nil)
	mockLogger.Debug("foo", nil)
	mockLogger.Notice("foo", nil)
	mockLogger.Err("foo", nil)
	mockLogger.Debug("foo", nil)

}

func TestMockLogFacility(t *testing.T) {
	testLogger := NewMockLogger()
	testLogger.facility = syslog.LOG_AUTH

	facility := testLogger.Facility()
	if facility != syslog.LOG_AUTH {
		t.Error("Facility priority does not match")
	}
}

func TestMockLogPriority(t *testing.T) {
	testLogger := NewMockLogger()
	testLogger.level = syslog.LOG_INFO

	facility := testLogger.Level()
	if facility != syslog.LOG_INFO {
		t.Error("Priority does not match")
	}
}

func TestMockLevelName(t *testing.T) {
	testLogger := NewMockLogger()

	err := testLogger.SetLevelName("BOGUS")
	if err == nil {
		t.Error("Expected an error for setting an invalid level name")
	}

	err = testLogger.SetLevelName("DEBUG")
	if err != nil {
		t.Error("Expected no error for setting a level name")
	}
}

func TestMockLevel(t *testing.T) {
	testLogger := NewMockLogger()

	err := testLogger.SetLevel(syslog.LOG_WARNING)
	if err != nil {
		t.Error("Expected no error for setting a level")
	}

	err = testLogger.SetLevel(-1)
	if err == nil {
		t.Error("Expected no error")
	}
}
