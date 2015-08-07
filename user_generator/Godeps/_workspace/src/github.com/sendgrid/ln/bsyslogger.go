package ln

import (
	"fmt"

	bsyslog "github.com/sendgrid/ln/vendor/blackjack"
)

// this struct wraps blackjack/syslog so that we can use it as a syslogger
type bSyslogger struct{}

//calls Openlog with facility and tag. In the blackjack library, this parameter is incorrectly named
//as priority when it should actually be facility because linux openlog takes facility only
func newBSyslogger(facility bsyslog.Priority, tag string) syslogger {
	bsyslog.Openlog(fmt.Sprintf("%s: %s", tag, tag), bsyslog.LOG_PID|bsyslog.LOG_CONS, facility)
	return &bSyslogger{}
}

func (_ *bSyslogger) Close() {
	bsyslog.Closelog()
}

func (b *bSyslogger) Emerg(msg string) error {
	bsyslog.Emerg(msg)
	return nil
}

func (b *bSyslogger) Alert(msg string) error {
	bsyslog.Alert(msg)
	return nil
}

func (b *bSyslogger) Crit(msg string) error {
	bsyslog.Crit(msg)
	return nil
}

func (b *bSyslogger) Err(msg string) error {
	bsyslog.Err(msg)
	return nil
}

func (b *bSyslogger) Notice(msg string) error {
	bsyslog.Notice(msg)
	return nil
}

func (b *bSyslogger) Warning(msg string) error {
	bsyslog.Warning(msg)
	return nil
}

func (b *bSyslogger) Info(msg string) error {
	bsyslog.Info(msg)
	return nil
}

func (b *bSyslogger) Debug(msg string) error {
	bsyslog.Debug(msg)
	return nil
}
