package mcredis

type Connection interface {
	Close() error
	Do(commandName string, args ...interface{}) (reply interface{}, err error)
}
