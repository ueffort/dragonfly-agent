package main

import (
	"fmt"
	"net"
	"os"
	"path"
	"time"

	"github.com/Sirupsen/logrus"
)
import "github.com/garyburd/redigo/redis"

type RedisServer struct {
	address   string
	options   []redis.DialOption
	prefix    string
	master    string
	conn      redis.Conn
	advertise *Advertise
}

var (
	logger *logrus.Logger
)

func init() {
	logrus.SetOutput(os.Stderr)
	logger = logrus.StandardLogger()
}

func (redisServer *RedisServer) init(discovery *Discovery, advertise *Advertise) error {
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
	redisServer.prefix = prefix[1:]
	redisServer.master = master
	err = redisServer.get()
	if err != nil {
		return err
	}
	redisServer.advertise = advertise
	return nil
}

func (redisServer *RedisServer) keep(second int) {
	logger.Infof("Keep every %s second", second)
	replay, err := redisServer.register()
	logger.Debugf("Register after %s second, replay->%s, err->%s", second, replay, err)
	go func() {
		for {
			select {
			case <-time.After(time.Second * time.Duration(second)):
				replay, err := redisServer.register()
				logger.Debugf("Register after %s second, replay->%s, err->%s", second, replay, err)
			}
		}
	}()
}

func (redisServer *RedisServer) register() (interface{}, error) {
	return redisServer.conn.Do("PUBLISH", redisServer.prefix+redisServer.master+"/"+DISCOVERY_REGISTER_NODE, redisServer.advertise.origin)
}

func (redisServer *RedisServer) watch() {

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
