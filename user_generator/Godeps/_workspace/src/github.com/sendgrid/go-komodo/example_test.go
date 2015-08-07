package komodo

import (
	"os"
)

type ExampleAdminable struct{}

type ExampleConfig struct{}

func (e *ExampleAdminable) Version() string {
	return "alpha"
}
func (e *ExampleAdminable) Name() string {
	return "example_app"
}

func (e *ExampleAdminable) Healthchecks() []Healthcheck {
	spoolCheck := &BasicHealthcheck{"SpoolDir", func() error {
		if _, err := os.Stat("/tmp"); err != nil {
			if os.IsNotExist(err) {
				return err
			}
		}
		return nil
	}}

	return []Healthcheck{
		spoolCheck,
	}
}

func (e *ExampleAdminable) MaintenanceFile() string {
	return "/tmp/healthcheck"
}

func (e *ExampleAdminable) Config() interface{} {
	return ExampleConfig{}
}

func Example() {
	adminable := &ExampleAdminable{}
	komodo := NewServer(adminable)

	// typically spawned in a goroutine
	komodo.ListenAndServe(":0")
}

type ExampleDebuggable struct {
	debug bool
}

func (e *ExampleDebuggable) Debug() bool {
	return e.debug
}

func (e *ExampleDebuggable) SetDebug(d bool) {
	e.debug = d
}

func ExampleDebuggableService() {
	adminable := &ExampleAdminable{}
	debuggable := &ExampleDebuggable{}

	// Adminable is a required interface to support komodo
	komodo := NewServer(adminable)

	// debugging support is optional
	komodo.Debuggable = debuggable
}
