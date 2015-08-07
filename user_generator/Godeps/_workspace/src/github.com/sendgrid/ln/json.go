package ln

import (
	"encoding/json"
	"errors"
	"log/syslog"
	"os"
)

type jsonLogger struct {
	s        syslogger
	level    syslog.Priority
	facility syslog.Priority
	callback func(error)
}

type message struct {
	Message string `json:"message"`
	Values  Map
}

func (j *jsonLogger) Close() {
	j.s.Close()
}

func newJSONLogger(s syslogger, level syslog.Priority, facility syslog.Priority) *jsonLogger {
	return &jsonLogger{s: s, level: level, facility: facility}
}

func (j *jsonLogger) SetErrorHandler(callback func(error)) {
	j.callback = callback
}

func (j *jsonLogger) SetLevel(l syslog.Priority) error {
	if l < 0 || l > syslog.LOG_DEBUG {
		return errors.New("invalid level " + string(l))
	}
	j.level = l

	return nil
}

func (j *jsonLogger) SetLevelName(l string) error {

	level, err := stringToLevel(l)
	if err != nil {
		return err
	}

	j.level = level

	return nil

}

func (j *jsonLogger) Level() syslog.Priority {
	return j.level
}

func (j *jsonLogger) Facility() syslog.Priority {
	return j.facility
}

func (j *jsonLogger) Emerg(msg string, v Map) error {
	return j.output(syslog.LOG_EMERG, j.s.Emerg, message{Message: msg, Values: v})
}

func (j *jsonLogger) Alert(msg string, v Map) error {
	return j.output(syslog.LOG_ALERT, j.s.Alert, message{Message: msg, Values: v})
}

func (j *jsonLogger) Crit(msg string, v Map) error {
	return j.output(syslog.LOG_CRIT, j.s.Crit, message{Message: msg, Values: v})
}

func (j *jsonLogger) Err(msg string, v Map) error {
	return j.output(syslog.LOG_ERR, j.s.Err, message{Message: msg, Values: v})
}

func (j *jsonLogger) Notice(msg string, v Map) error {
	return j.output(syslog.LOG_NOTICE, j.s.Notice, message{Message: msg, Values: v})
}

func (j *jsonLogger) Warning(msg string, v Map) error {
	return j.output(syslog.LOG_WARNING, j.s.Warning, message{Message: msg, Values: v})
}

func (j *jsonLogger) Info(msg string, v Map) error {
	return j.output(syslog.LOG_INFO, j.s.Info, message{Message: msg, Values: v})
}

func (j *jsonLogger) Debug(msg string, v Map) error {
	return j.output(syslog.LOG_DEBUG, j.s.Debug, message{Message: msg, Values: v})
}

func (j *jsonLogger) Fatal(msg string, v Map) {
	j.output(syslog.LOG_EMERG, j.s.Emerg, message{Message: msg, Values: v})
	os.Exit(1)
}

// used in logging an event without a message
// always use INFO for Event logging.
const EVENT_LOG_LEVEL = syslog.LOG_INFO

func (j *jsonLogger) Event(event string, v Map) error {
	// we should honor the event that's passed in explictly
	v["event"] = event
	return j.output(EVENT_LOG_LEVEL, j.s.Info, message{Values: v})
}

func (j *jsonLogger) output(level syslog.Priority, output func(v string) error, m message) error {
	if j.level < level {
		return nil
	}

	logMap := m.Values

	if logMap == nil {
		logMap = make(Map)
	}

	// no point of logging message if it's empty
	if m.Message != "" {
		logMap["message"] = m.Message
	}

	b, err := json.Marshal(logMap)
	if err != nil {
		return errors.New("could not marshal json: " + err.Error())
	}

	err = output(string(b))

	if err != nil {

		if j.callback != nil {
			j.callback(err)
		}
		return err
	}
	return nil
}
