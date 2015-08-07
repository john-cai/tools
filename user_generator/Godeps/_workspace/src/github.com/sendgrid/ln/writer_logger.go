package ln

import (
	"fmt"
	"io"
	"log"
)

type writerLogger struct {
	l *log.Logger
}

const LoggerCallStackDepth = 6

func (_ *writerLogger) Close() {}

func newWriterLogger(w io.Writer, tag string) syslogger {
	return &writerLogger{
		l: log.New(w, fmt.Sprintf("[%s] ", tag), log.LstdFlags|log.Lshortfile),
	}
}

func (s *writerLogger) output(level string, msg string) {
	s.l.Output(LoggerCallStackDepth, "["+level+"] "+msg)
}

func (s *writerLogger) Emerg(msg string) error {
	s.output("emerg", msg)
	return nil
}

func (s *writerLogger) Alert(msg string) error {
	s.output("alert", msg)
	return nil
}

func (s *writerLogger) Crit(msg string) error {
	s.output("crit", msg)
	return nil
}

func (s *writerLogger) Err(msg string) error {
	s.output("err", msg)
	return nil
}

func (s *writerLogger) Notice(msg string) error {
	s.output("notice", msg)
	return nil
}

func (s *writerLogger) Warning(msg string) error {
	s.output("warning", msg)
	return nil
}

func (s *writerLogger) Info(msg string) error {
	s.output("info", msg)
	return nil
}

func (s *writerLogger) Debug(msg string) error {
	s.output("debug", msg)
	return nil
}
