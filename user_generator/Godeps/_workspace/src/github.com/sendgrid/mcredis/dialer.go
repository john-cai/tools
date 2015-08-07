package mcredis

import "time"

type Dialer interface {
	Dial(connectionType string, address string) (Connection, error)
	DialTimeout(
		connectionType string,
		address string,
		connectTimeout,
		readTimeout,
		writeTimeout time.Duration,
	) (Connection, error)
}
