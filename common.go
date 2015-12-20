package main

import (
	"fmt"
	"net"
	"os"
	"path"
	"time"

	"github.com/Sirupsen/logrus"
	redisgo "github.com/garyburd/redigo/redis"
)

type Redis struct {
	address string
	options []redisgo.DialOption
	prefix  string
	master  string
}

var (
	logger *logrus.Logger
)

func init() {
	logrus.SetOutput(os.Stderr)
	logger = logrus.StandardLogger()
}

func (redis *Redis) init(discovery *Discovery) error {
	host, port, err := net.SplitHostPort(discovery.url.Host)
	if err != nil {
		host = discovery.url.Host
	}
	if host == "" {
		return fmt.Errorf("RedisServer host is empty: %s", discovery.url.Host)
	}
	if port == "" {
		port = "6379"
	}
	logger.Debugf("RedisServer info: host->%s, port->%s", host, port)
	address := net.JoinHostPort(host, port)
	redis.address = address
	password, isSet := discovery.url.User.Password()
	if isSet {
		redis.options = append(redis.options, redisgo.DialPassword(password))
	}
	prefix, master := path.Split(discovery.url.Path)
	if prefix == "" {
		return fmt.Errorf("Redis path is empty: %s", discovery.url.Path)
	}
	if master == "" {
		master = "master"
	}
	redis.prefix = prefix[1:]
	redis.master = master
	return nil
}

func (redis *Redis) notice(target string, message interface{}) (interface{}, error) {
	conn, err := redis.connect()
	defer conn.Close()
	if err != nil {
		return nil, err
	}
	return conn.Do("PUBLISH", redis.prefix+redis.master+"/"+target, message)
}

func (redis *Redis) watch(advertise string, handle, watching chan<- bool, unwartch <-chan bool) {
	conn, err = redis.connect()
	defer conn.Close()
	if err != nil {
		return nil, err
	}
	psc := redisgo.PubSubConn{Conn: conn}
	psc.Subscribe(redis.prefix + advertise)
	watching <- bool
	for {
		switch {
		case <-unwartch:
			psc.Unsubscribe()
			psc.PUnsubscribe()
			return
		default:
			v := psc.Receive()
			switch v.(type) {
			case redisgo.Message:
				fallthrough
			case redisgo.PMessage:
				handle(v.Data)
			case error:
				return v
			}
		}
	}
}

func (redis *Redis) connect() (redisgo.Conn, error) {
	if redis.address == "" {
		return fmt.Errorf("Redis not set, setRedis first:%s", redis)
	}
	return redisgo.Dial("tcp", redis.address, redis.options...)
}
