package main

import (
	"fmt"
	"net"
	"os"
	"path"

	"github.com/Sirupsen/logrus"
)
import "github.com/garyburd/redigo/redis"

type RedisServer struct {
	address string
	options []redis.DialOption
	prefix  string
	master  string
	conn    redis.Conn
}

var (
	logger *logrus.Logger
)

func init() {
	logrus.SetOutput(os.Stderr)
	logger = logrus.StandardLogger()
}

func (redisServer *RedisServer) init(discovery *Discovery) error {
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
	redisServer.address = address
	password, isSet := discovery.url.User.Password()
	if isSet {
		redisServer.options = append(redisServer.options, redis.DialPassword(password))
	}
	prefix, master := path.Split(discovery.url.Path)
	if prefix == "" {
		return fmt.Errorf("RedisServer path is empty: %s", discovery.url.Path)
	}
	if master == "" {
		master = "master"
	}
	redisServer.prefix = prefix
	redisServer.master = master
	err = redisServer.get()
	if err != nil {
		return err
	}
	return nil
}

func (redisServer *RedisServer) register(advertise *Advertise) {
	replay, err := redisServer.conn.Do("APPEND", "key", "value")
	if err != nil {
		panic(err)
	}
	replay, err = redisServer.conn.Do("PUBLISH", "dragonfly/master", 123)
	if err != nil {
		panic(err)
	}
	logger.Debugln(replay)
	//	redisServer.conn.Send("PUBLISH", redisServer.prefix+"/"+redisServer.master, advertise.origin)
}

func (redisServer *RedisServer) watch(advertise *Advertise) {

}

func (redisServer *RedisServer) get() error {
	if redisServer.address == "" {
		return fmt.Errorf("Redis not set, setRedis first:%s", redisServer)
	}
	conn, err := redis.Dial("tcp", redisServer.address, redisServer.options...)
	if err != nil {
		return err
	}
	redisServer.conn = conn
	return nil
}
