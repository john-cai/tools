package ln

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestWriterLogger(t *testing.T) {
	b := bytes.NewBuffer(nil)
	tag := "testtag"
	l := newWriterLogger(b, tag)

	testCases := []struct {
		name   string
		method func(string) error
	}{
		{"emerg", l.Emerg},
		{"alert", l.Alert},
		{"crit", l.Crit},
		{"err", l.Err},
		{"notice", l.Notice},
		{"warning", l.Warning},
		{"info", l.Info},
		{"debug", l.Debug},
	}

	for _, c := range testCases {
		expected := c.name
		c.method("dummy message")
		actual := b.String()
		b.Reset()
		if !strings.HasPrefix(actual, fmt.Sprintf("[%s]", tag)) || !strings.Contains(actual, expected) {
			t.Errorf("writerLogger.%s got %q, expected %q", c.name, actual, expected)
		}
	}
}
