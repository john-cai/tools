package ln

import (
	"flag"
	"io"
	"log/syslog"
)

var endPoint = flag.String("ln_output", "syslog", "syslog|stderr")
var facility = flag.String("ln_facility", "LOCAL0", "Typically LOCAL0. Refer to syslog documentation: http://en.wikipedia.org/wiki/Syslog#Facility_levels")
var level = flag.String("ln_level", "INFO", "The lowest severity level to be logged. EMERG|ALERT|CRIT|ERR|NOTICE|WARNING|INFO|DEBUG")
var tag = flag.String("ln_tag", "application", "Identifier with which your logs are tagged.")

// The current thought for this function is that it is useful for tests to suppress or capture logs
func SetOutput(w io.Writer, tag string) {
	baseLogger = newJSONLogger(newWriterLogger(w, tag), syslog.LOG_DEBUG, -1)
}

func Fatal(msg string, v Map) {
	baseLogger.Fatal(msg, v)
}

func Emerg(msg string, v Map) error {
	return baseLogger.Emerg(msg, v)
}

func Alert(msg string, v Map) error {
	return baseLogger.Alert(msg, v)
}

func Crit(msg string, v Map) error {
	return baseLogger.Crit(msg, v)
}

func Err(msg string, v Map) error {
	return baseLogger.Err(msg, v)
}

func Notice(msg string, v Map) error {
	return baseLogger.Notice(msg, v)
}

func Warning(msg string, v Map) error {
	return baseLogger.Warning(msg, v)
}

func Info(msg string, v Map) error {
	return baseLogger.Info(msg, v)
}

func Debug(msg string, v Map) error {
	return baseLogger.Debug(msg, v)
}

func Level() syslog.Priority {
	return baseLogger.Level()
}

func SetLevel(l syslog.Priority) error {
	return baseLogger.SetLevel(l)
}

func SetLevelName(l string) error {
	return baseLogger.SetLevelName(l)
}
