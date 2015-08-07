package envy

import (
	"os"
	"strings"
)

// interface that reads config from somewhere
type EnvironmentReader interface {
	// Method reads the environment from the source
	//
	// Returns: map[string]string of environment keys to values
	Read() map[string]string

	// Method returns the prefix for more helpful logging
	GetPrefix() string
}

// Default EnvironmentReader
// Reads environment with the provided prefix, defaulted to ""
type OsEnvironmentReader struct {
	Prefix string
	Source func() []string
}

// Reads values from the os.Environ slice and returns the result
// as a map[string]string
func (o *OsEnvironmentReader) Read() map[string]string {
	// default to use os.Environ: Not Testable
	if o.Source == nil {
		o.Source = os.Environ
	}
	result := make(map[string]string)
	for _, envVar := range o.Source() {
		if strings.HasPrefix(envVar, o.Prefix) {
			parts := strings.SplitN(envVar, "=", 2)

			// remove the prefix so we don't have to use it on the provided struct
			key := strings.TrimPrefix(parts[0], o.Prefix)
			value := parts[1]
			result[key] = value
		}
	}

	return result
}

func (o *OsEnvironmentReader) GetPrefix() string {
	return o.Prefix
}
