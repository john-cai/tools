package mcredis

import (
	"fmt"
	"strings"
	"time"

	"github.com/sendgrid/waitfor"

	"github.com/garyburd/redigo/redis"
)

type RedisSentinel struct {
	servers     []string
	masterHost  string
	masterPort  string
	masterName  string
	callTimeout time.Duration
	timeout     time.Duration
	connPool    *redis.Pool
	poolSize    int
	maxIdle     int
	dialer      RedisDialer
}

func NewSentinel(conf RedisConfig) (*RedisSentinel, error) {
	rs := &RedisSentinel{
		servers:     conf.Nodes,
		callTimeout: conf.CallTimeout,
		timeout:     conf.Timeout,
		poolSize:    conf.PoolSize,
		maxIdle:     conf.MaxIdle,
		dialer:      McRedisDialer{},
		masterName:  conf.MasterName,
	}
	err := rs.initialize()
	if err != nil {
		return nil, err
	}
	return rs, nil
}

func (r *RedisSentinel) initialize() error {
	var err error
	r.masterHost, err = r.connectAndGetMaster()
	if err != nil {
		return err
	}
	r.connPool = r.newPool()
	return nil
}

// Use this function to rebuild the connection schematic if the master changes or the connection pool borks
func (r *RedisSentinel) RebuildConnectionPool() error {
	return r.initialize()
}

func (r *RedisSentinel) newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     r.maxIdle,
		IdleTimeout: r.timeout,
		MaxActive:   r.poolSize,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", r.masterHost)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func (r *RedisSentinel) Do(action string, params ...interface{}) (result interface{}, err error) {
	var conn Connection
	err = waitfor.Func(func() (err error) {
		conn, err = r.getConn()
		return
	}, time.Duration(1)*time.Millisecond, r.callTimeout)
	if err != nil {
		return
	}

	defer conn.Close()
	err = waitfor.Func(func() (err error) {
		result, err = conn.Do(action, params...)
		return
	}, time.Duration(1)*time.Millisecond, r.callTimeout)
	if err != nil {
		if strings.Contains(err.Error(), "You can't write against a read only slave.") { // this is very hacky
			err = r.RebuildConnectionPool()
			if err != nil {
				return
			}
			result, err = r.Do(action, params...)
		} else {
			return
		}
	}
	return
}

func (r *RedisSentinel) getConn() (Connection, error) {
	return r.connPool.Get(), nil
}

func (r *RedisSentinel) connectAndGetMaster() (string, error) {

	for _, server := range r.servers {
		if !strings.Contains(server, ":") {
			return "", MalformedServerNodeError{"Node: " + server + " does not contain a port"}
		}
		initialConn, err := r.dialer.DialTimeout("tcp", server, 300*time.Millisecond, 0, 0)
		if err != nil {
			break
		}
		defer initialConn.Close()

		masterAddr, err := r.getMasterAddressGivenClusterConnection(initialConn)
		if err != nil {
			break
		}

		if r.confirmMasterConnection(masterAddr) {
			return masterAddr, nil
		}
	}
	return "", NoSentinelFound{"Failed to connect to a sentinel"}
}

// GetInfo() returns the master address and the address of the sentinel that it connected to
func (r *RedisSentinel) GetInfo() (string, string) {
	for _, server := range r.servers {
		sentinelConn, err := r.dialer.DialTimeout("tcp", server, 300*time.Millisecond, 0, 0)
		if err != nil {
			continue
		}
		defer sentinelConn.Close()
		masterAddr, err := sentinelConn.Do("SENTINEL", "get-master-addr-by-name", r.masterHost)
		if err != nil {
			continue
		}
		return constructMasterAddress(masterAddr), server
	}
	return "", ""
}

func (r *RedisSentinel) getMasterAddressGivenClusterConnection(clusterConn Connection) (string, error) {

	if r.masterName == "" {
		return "", NoMasterNameError{"You must include a MasterName in your sentinel config"}
	}

	var masterAddr interface{}
	masterAddr, err := clusterConn.Do("SENTINEL", "get-master-addr-by-name", r.masterName)
	if err != nil {
		return "", err
	}
	if masterAddr == nil {
		return "", MasterNameNotKnownError{fmt.Sprintf("Sentinel does not know the master: %s", r.masterName)}
	}

	return constructMasterAddress(masterAddr), nil
}

func constructMasterAddress(masterAddr interface{}) string {
	masterAddrSlice := masterAddr.([]interface{})
	masterAddrIP, _ := redis.String(masterAddrSlice[0], nil)
	masterAddrPort := "6379" // hardcoding this for now, we need to figure out a dynamic way to get the master's Sentinel and Command ports

	return masterAddrIP + ":" + masterAddrPort
}

func (r *RedisSentinel) confirmMasterConnection(masterAddr string) bool {
	masterConn, err := r.dialer.Dial("tcp", masterAddr)
	if err != nil {
		return false
	}

	masterInfo, err := masterConn.Do("info", "replication")
	masterInfoString, _ := redis.String(masterInfo, err)
	if !strings.Contains(masterInfoString, "role:master") {
		return false
	}
	return true
}
