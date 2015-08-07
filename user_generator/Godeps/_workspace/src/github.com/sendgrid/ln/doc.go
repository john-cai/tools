/*
Usage:

	// syslog
	var logger ln.LevelLogger
	logger = ln.New("syslog", "debug", "auth", "test app new")
	logger.Alert("Test logging message to SYSLOG", ln.Map{"k1": "some v1", "k2": "some v2"})
	logger.Info("Test logging message ** to SYSLOG", nil)

	// a "loggable" struct
	ts := testStruct{"hidden", "public"}

	// default values
	logger = ln.New("syslog", "", "", "test app new")
	logger.Info("Test logging message ** to SYSLOG using default", ln.Map{"testStruct": &ts})

	// stderr
	logger = ln.New("stderr", "", "", "test app new")
	logger.Info("Test logging message ** to STDERR", nil)


Note: logger.Fatal() acts like log.Fatal() in Go and kills the process.  Example below:

	// logging .Fatal should kill the process after writing the log message to Emerg

	go func() {
		for {
			<-time.After(1 * time.Second)
			fmt.Println("I'm still running!")
		}
	}()

	<-time.After(3 * time.Second)
	logger.Fatal("DIE!", nil)  // this will print out "I'm still running!" twice
}
*/
package ln
