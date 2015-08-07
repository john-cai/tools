# ln

A more natural approach to structured logging in go. Conforms to the SendGrid [logging archetype](https://wiki.sendgrid.net/display/SD/Logging+Design)

### Usage ###

Configurables:
 - LN_OUTPUT     = STDERR | SYSLOG | &#91;filename&#93; (STDERR and SYSLOG are case insensitive and the default is syslog)
 - LN_FACILITY   = Defaults to LOCAL0. Refer to syslog documentation: http://en.wikipedia.org/wiki/Syslog#Facility_levels
 - LN_LEVEL      = The lowest severity level to be logged. EMERG | ALERT | CRIT | ERR | NOTICE | WARNING | INFO | DEBUG
 - LN_TAG        = Identifier with which your logs are tagged (usually an application name)

Ln respects OS ENV vars. Below are examples of passing configuration into the ```ln``` package.

```
# on cmd line
$ go run main.go                                   # Logs to syslog (by default)
$ LN_OUTPUT=stderr go run main.go                  # Logs to stderr
$ LN_OUTPUT=my.log go run main.go                  # logs to my.log
$ LN_TAG="my app" LN_OUTPUT=my.log go run main.go  # logs to my.log with a tag of [my app]
```

In your main, you can use the logger as you see below. The base package logger will respect the configs passed in via environment variables.

```go
package main

import ("github.com/sendgrid/ln")

func main(){
	// You should user either the package level logger OR instantiate your own logger.
	// Mixing them can lead to syslog connection issues.

	// Option 1: using the package level logger
	ln.SetTag("application")
	ln.Err("Error 1", nil)

	// Option 2: using instantiated logger. If you pass a string, like 'stderr', then it will always log that way.
	// You can have it read in the configuration (env var) to make it runtime-customizable.
	logger := ln.New("stderr", "DEBUG", "LOCAL0", "application")
	logger.Err("Error 2", nil)

	// NOT AN OPTION: Don't do this. If you create a package level logger
	// and then create an instantiated logger, and then, again, create a package
	// level logger, then we've now lost the syslog connection by making a new stderr logger.
	// Avoid this by picking option 1 or 2 and sticking with it. This can be fixed with a major
	// version update.
	ln.Err("Error 3", nil)
}
```

### Tests ###

For tests, you can either direct your log output to a file using the LN_OUTPUT environment variable. Alternatively, you can assign an `io.Writer` to your logger to capture logs. using `func SetOutput(w io.Writer, tag string)`. Note, this affects the internal baseLogger, not any logger that your service instantiates itself with `New()`.

```go
package main

import (
	"github.com/sendgrid/ln"
	"os"
	"testing"
)

func TestIt(t *testing.T) {
	// create some io.Writer
	file, err := os.Create("log.log")
	if err != nil {
		t.Fatal("Can't open log file!")
	}

	// Set with an io.Writer and a tag
	ln.SetOutput(file, "testing")

	// now any logging that happens in SomeFuncToTest() is captured in the io.Writer
	// and you don't need to worry about stderr logs clogging up your terminal.
	if SomeFuncToTest() != 1 {
		t.Error("Errors!")
	}
}
```

### Current Limitations ###

Keep in mind that as a syslog utility, that you may have DEBUG and other lower level messages supressed in syslog. Additionally, you may have EMERG level messages broadcast to multiple terminal sessions.

1.1.1 - We introduced package level functions. We are unable to use the package level logger, switch to the embedded logger, and then switch back to the package level logger. This causes the connection to syslog to become permanently closed.

1.1.0 - We are using the blackjack library, a c-bindings implementation of syslog(). The way it's implemented is through a global call to C's syslog(). We didn't want to change our implementation to use package level functions as logger calls, so then we are enforcing LevelLogger to be a singleton.

That means only one instance of logger can exist per application


### Run Tests ###
```
    GOPATH=$PWD go test
```
To generate code coverage reports, run this:
```
    go test -coverprofile=cover.out
    sed -i -e "s^_$PWD^.^" cover.out
    go tool cover -html=cover.out
```


### Default Settings

The code will set the following defaults:

- ln_output defaults to syslog
- ln_tag defaults to the string 'ln'
- ln_facility defaults to 'LOCAL0'
- ln_level defaults to 'DEBUG'

To override these, you may set ENV variables as such:
```
export LN_OUTPUT=stderr
export LN_TAG=go_segment
export LN_FACILITY=LOCAL0
export LN_LEVEL=DEBUG
```

Once you start your application, you can can override the 'tag' (application name) in your code by calling the SetTag() method. This is NOT for instantiated loggers (using New())

```
import (
  "github.com/sendgrid/ln"
)

func main() {
    ln.Err("my string")
    // this would output an application name of [ln] as the first field in the log output
    
    
    ln.SetTag("my app")
    ln.Err("my other string")
    // this would output an application name of [my app] as the first field in the log output
}
```
