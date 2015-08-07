package ln

import (
	"encoding/json"
	"log/syslog"
	"os"
	"testing"

	bsyslog "github.com/sendgrid/ln/vendor/blackjack"
)

// since we enforce a singleton on bSysLogger, for testing purposes we need to "destroy" the singleton
// so we can instantiate a new one in another test

func destroyBSyslog() {
	bsyslog.Closelog()
}

// Test that the New constructor panics when called with an invalid endpoint
func TestInvalidLogOutputFile(t *testing.T) {
	// if endpoint isn't syslog or stderr, should give an error
	defer func() {
		destroyBSyslog()
		if r := recover(); r != nil {

			if r != "Initializing ln logger: could not create endpoint readonly.file, error: open readonly.file: permission denied" {
				t.Errorf("should have panicked with message 'Initializing ln logger: could not create endpoint readonly.file, error: open readonly.file: permission denied', got %s", r)
			}
			return
		}

		t.Errorf("should have panicked")
	}()

	_, err := os.OpenFile("readonly.file", os.O_CREATE|os.O_RDONLY, 0444)
	if err != nil {
		t.Fatal("Unable to open file")
	}
	defer os.Remove("readonly.file")
	New("readonly.file", "emerg", "auth", "test app")

}

// Test that the New constructor creates a new log file properly
func TestCreationOfLogOutputFile(t *testing.T) {
	defer os.Remove("new.log")
	New("new.log", "emerg", "auth", "test app")
	_, err := os.Open("new.log")
	if err != nil {
		t.Error("Log file not created")
	}
}

// Test that the New constructor creates a new log file properly
func TestAppendOfLogOutputFile(t *testing.T) {
	defer os.Remove("appended.log")
	logger := New("appended.log", "DEBUG", "auth", "test app")
	_, err := os.Open("appended.log")
	if err != nil {
		t.Error("Log file not created")
	}

	logger.Err("foo", nil)
	fh, _ := os.Open("appended.log")
	fileinfo, _ := fh.Stat()
	size1 := fileinfo.Size()

	logger.Err("bar", nil)
	fileinfo, _ = fh.Stat()
	size2 := fileinfo.Size()

	if size2 <= size1 {
		t.Error("Log file append didn't change log file size")
	}
}

// Test that the New constructor panics when called with an invalid logging level
func TestNewErrorWhenInvalidLevel(t *testing.T) {

	defer func() {
		destroyBSyslog()
		if r := recover(); r != nil {

			if r != "Initializing in ln logger: unrecognized syslog level nonsense" {
				t.Errorf("should have panicked with message 'Initializing in ln logger: unrecognized syslog level nonsense', got %s", r)

			}
			return
		}

		t.Errorf("should have panicked")

	}()

	New("syslog", "nonsense", "auth", "test app")
}

// Test that the New constructor panics when called with an invalid logging facility
func TestNewErrorWhenInvalidFacitlity(t *testing.T) {

	defer func() {
		destroyBSyslog()
		if r := recover(); r != nil {

			if r != "Initializing in ln logger : unrecognized syslog facility nowhere" {
				t.Errorf("should have panicked with message 'Initializing in ln logger : unrecognized syslog facility nowhere', got %s", r)

			}
			return
		}

		t.Errorf("should have panicked")

	}()

	New("syslog", "info", "nowhere", "test app")
}

// Test that the New constructor gives a valid LevelLogger that is set to the correct logging level
func TestNew(t *testing.T) {
	logger := New("syslog", "info", "auth", "test app")
	defer destroyBSyslog()

	_, ok := logger.(LevelLogger)
	if !ok {
		t.Error("New() should return a LevelLogger")
	}
	if logger.Level() != syslog.LOG_INFO {
		t.Error("level returned from Level() should match what's passed into New()")
	}
}

// This is how you would make a struct "loggable"
// we need to make sure the struct that is passed to the logger should be JSON-Marshallable
type testStruct struct {
	someField     string
	ExportedField string
}

func (t *testStruct) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"somefield":     t.someField,
		"ExportedField": t.ExportedField,
	})
}

// Example of how to create a new logger and use it
func ExampleNew() {
	defer destroyBSyslog()

	logger := New("syslog", "debug", "auth", "test app new")
	logger.Alert("Test logging message to SYSLOG", Map{"k1": "some v1", "k2": "some v2"})
	logger.Info("Test logging message ** to SYSLOG", nil)
}

// Example of how log an arbitrary struct with unexported fields, requiring its own MarshalJSON() func
func ExampleNewDefault() {
	defer destroyBSyslog()

	// a "loggable" struct
	ts := testStruct{"hidden", "public"}

	// default values
	logger := New("syslog", "", "", "test app new")
	logger.Info("Test logging message ** to SYSLOG using default", Map{"testStruct": &ts})
}

// Example of creating a stderr logger
func ExampleNewStderr() {
	defer destroyBSyslog()

	logger := New("stderr", "", "", "test app new")
	logger.Info("Test logging message ** to STDERR", nil)
}
