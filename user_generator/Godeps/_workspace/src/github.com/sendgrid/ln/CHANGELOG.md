Change log
==========
## 1.10.0
- Add Merge function

## 1.9.2
- Revert duplicate tag fix as it is not needed with removal of logstash

## 1.9.1
- Fix duplicate tag issue (this will not be recognized by Splunk)

## 1.9.0
- Add EventLevelLogger interface.  Add Event() to jsonLogger.

## 1.8.0
- add Null Logger

## 1.7.1
- Use mutex for mock logger.

## 1.7.0
- Replaced mock.go AssertLogged() to use a test interface instead of standard golang default testing struct

## 1.6.2
- Change writer_logger output to include correct file name and line numbers for development

## 1.6.1
- Add Fatal

## 1.6.0
- Added Level(), SetLevel() and SetLevelName() for the baseLogger.

## 1.5.0
- Change the tag passed into blacjack constructor to be "appname: appname". This is to enable our rsyslog pattern matchers to match on the appname.

## 1.4.0
- Removed flag.Parse() from init, set up some defaults where ENV is not set
- allow applications to call a SetTag() method to alter logger 'tag' for the first field in log file output

## 1.3.2
- Make bSyslogger threadsafe by removing the singleton pattern

## 1.3.1
- MockLogger.exited changed to MockLogger.Exited so that external packages can make use of MockLogger.Fatal()

## 1.3.0
- services can log to file

## 1.2.0
- package level functions for all log levels

## 1.1.0
- use blackjack/syslog in place of log/syslog (with wrapper struct bsyslogger)

## 1.0.0
- initial library interface
