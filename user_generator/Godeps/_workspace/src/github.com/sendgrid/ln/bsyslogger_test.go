package ln

import (
	"log/syslog"
	"testing"

	bsyslog "github.com/sendgrid/ln/vendor/blackjack"
)

func TestSysloggerLevels(t *testing.T) {

	levels := []string{"EMERG", "ALERT", "CRIT", "WARNING", "INFO", "DEBUG", "NOTICE", "ERR"}
	facilities := []string{"KERN", "USER", "MAIL", "DAEMON", "AUTH", "SYSLOG", "LPR", "NEWS", "UUCP", "CRON", "AUTHPRIV", "FTP", "LOCAL0", "LOCAL1", "LOCAL2", "LOCAL3", "LOCAL4", "LOCAL5", "LOCAL6", "LOCAL7"}

	var bsl LevelLogger

	for _, level := range levels {
		New("syslog", level, "LOCAL0", "tag")
	}
	for _, facility := range facilities {
		bsl = New("syslog", "DEBUG", facility, "tag")
	}

	// bsl.Emerg("foo", nil) // commented out so it doesn't spam/error out on chef runs, but it passes, really!
	bsl.Alert("foo", nil)
	bsl.Crit("foo", nil)
	bsl.Warning("foo", nil)
	bsl.Info("foo", nil)
	bsl.Debug("foo", nil)
	bsl.Notice("foo", nil)
	bsl.Err("foo", nil)
	bsl.Debug("foo", nil)
}

func ExampleNewBSyslogger() {
	priority := bsyslog.Priority(syslog.LOG_CRIT | syslog.LOG_AUTH)
	syslogger := newBSyslogger(priority, "example app")
	syslogger.Info("hello world")
}
