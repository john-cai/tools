package ln

import (
	"errors"
	"log/syslog"
	"os"
	"strings"

	bsyslog "github.com/sendgrid/ln/vendor/blackjack"
)

// convenience type for arbitrary maps
type Map map[string]interface{}

// Logger that looks and acts like syslog
type syslogger interface {
	Close()
	Emerg(string) error
	Alert(string) error
	Crit(string) error
	Err(string) error
	Notice(string) error
	Warning(string) error
	Info(string) error
	Debug(string) error
}

// Level-based logger based on the log/syslog interface
type LevelLogger interface {
	Close()
	Emerg(message string, v Map) error
	Alert(message string, v Map) error
	Crit(message string, v Map) error
	Err(message string, v Map) error
	Notice(message string, v Map) error
	Warning(message string, v Map) error
	Info(message string, v Map) error
	Debug(message string, v Map) error
	Fatal(message string, v Map)
	SetErrorHandler(callback func(error))
	SetLevel(l syslog.Priority) error
	SetLevelName(l string) error
	Level() syslog.Priority
	Facility() syslog.Priority
}

// LevelLogger that also can log an event without a message
type EventLevelLogger interface {
	LevelLogger
	Event(event string, v Map) error
}

var baseLogger LevelLogger

func setVals() (*string, *string, *string, *string) {
	setVal("ln_output", endPoint, "syslog")
	setVal("ln_level", level, "DEBUG")
	setVal("ln_facility", facility, "LOCAL0")
	setVal("ln_tag", tag, "ln")
	return endPoint, level, facility, tag
}

func init() {
	endPoint, level, facility, tag = setVals()
	baseLogger = newLogger(*endPoint, *level, *facility, *tag)
}

// Updates logger tag for a new application name
//
// Args:
//   - tag: string of an application name
func SetTag(tag string) {
	endPoint, level, facility, _ = setVals()
	baseLogger = newLogger(*endPoint, *level, *facility, tag)
}

// Takes flag values over os ENV values
func setVal(s string, flagVar *string, default_setting string) {
	if val := os.Getenv(strings.ToUpper(s)); val != "" {
		*flagVar = val
	} else {
		*flagVar = default_setting
	}
}

// Creates a new Logger
//
// Args:
//   - output: either "syslog" or "stderr" or "filename"
//   - level: logging level "EMERG", "ALERT", "CRIT", "ERR", "WARNING", "NOTICE", "INFO", "DEBUG"
//   - facility: logging facility "KERN", "USER", "MAIL", "DAEMON", "AUTH", "SYSLOG", "LPR", "NEWS", "UUCP"
//                               ,"CRON", "AUTHPRIV", "FTP", "LOCAL0", "LOCAL1", "LOCAL2", "LOCAL3", "LOCAL4"
//                               ,"LOCAL5", "LOCAL6", "LOCAL7"
//         special case: "LOCAL0" will get sent to Splunk
//   - tag: Unique identifier. Typically the application names
func New(output, level, facility, tag string) LevelLogger {
	baseLogger.Close()
	return newLogger(output, level, facility, tag)
}

func NewEventLogger(output, level, facility, tag string) EventLevelLogger {
	baseLogger.Close()
	return newLogger(output, level, facility, tag)
}

func newLogger(endpoint, strLevel, strFacility, tag string) EventLevelLogger {
	// figure priority
	level, err := stringToLevel(strLevel)

	if err != nil {
		panic("Initializing in ln logger: " + err.Error())
	}

	// figure out endpoint
	if strings.ToLower(endpoint) == "stderr" {
		// create a logger that writes to stderr, and we dont' care about the facility
		return newJSONLogger(newWriterLogger(os.Stderr, tag), level, -1)
	} else if strings.ToLower(endpoint) == "syslog" {
		facility, err := stringToFacility(strFacility)

		if err != nil {
			panic("Initializing in ln logger : " + err.Error())
		}

		syslogger := newBSyslogger(bsyslog.Priority(facility), tag)
		return newJSONLogger(syslogger, level, facility)
	} else {
		file, err := os.OpenFile(endpoint, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			panic("Initializing ln logger: could not create endpoint " + endpoint + ", error: " + err.Error())
		}
		return newJSONLogger(newWriterLogger(file, tag), level, -1)
	}
}

// Takes a variadic number of maps and merges them together. Later keys override earlier keys
func Merge(maps ...Map) Map {
	newMap := Map{}
	for _, currMap := range maps {
		for k, v := range currMap {
			newMap[k] = v
		}
	}

	return newMap
}

func stringToLevel(syslogLevel string) (syslog.Priority, error) {
	switch strings.ToUpper(syslogLevel) {
	case "EMERG":
		return syslog.LOG_EMERG, nil
	case "ALERT":
		return syslog.LOG_ALERT, nil
	case "CRIT":
		return syslog.LOG_CRIT, nil
	case "ERR":
		return syslog.LOG_ERR, nil
	case "NOTICE":
		return syslog.LOG_NOTICE, nil
	case "WARNING":
		return syslog.LOG_WARNING, nil
	case "INFO":
		return syslog.LOG_INFO, nil
	case "DEBUG":
		return syslog.LOG_DEBUG, nil
	default:
		return -1, errors.New("unrecognized syslog level " + syslogLevel)
	}
}

func stringToFacility(syslogFacility string) (syslog.Priority, error) {
	switch strings.ToUpper(syslogFacility) {
	case "KERN":
		return syslog.LOG_KERN, nil
	case "USER":
		return syslog.LOG_USER, nil
	case "MAIL":
		return syslog.LOG_MAIL, nil
	case "DAEMON":
		return syslog.LOG_DAEMON, nil
	case "AUTH":
		return syslog.LOG_AUTH, nil
	case "SYSLOG":
		return syslog.LOG_SYSLOG, nil
	case "LPR":
		return syslog.LOG_LPR, nil
	case "NEWS":
		return syslog.LOG_NEWS, nil
	case "UUCP":
		return syslog.LOG_UUCP, nil
	case "CRON":
		return syslog.LOG_CRON, nil
	case "AUTHPRIV":
		return syslog.LOG_AUTHPRIV, nil
	case "FTP":
		return syslog.LOG_FTP, nil
	case "LOCAL0":
		return syslog.LOG_LOCAL0, nil
	case "LOCAL1":
		return syslog.LOG_LOCAL1, nil
	case "LOCAL2":
		return syslog.LOG_LOCAL2, nil
	case "LOCAL3":
		return syslog.LOG_LOCAL3, nil
	case "LOCAL4":
		return syslog.LOG_LOCAL4, nil
	case "LOCAL5":
		return syslog.LOG_LOCAL5, nil
	case "LOCAL6":
		return syslog.LOG_LOCAL6, nil
	case "LOCAL7":
		return syslog.LOG_LOCAL7, nil
	default:
		return -1, errors.New("unrecognized syslog facility " + syslogFacility)
	}
}
