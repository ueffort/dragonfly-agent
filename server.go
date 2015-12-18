package main

import (
	"fmt"
	"net/url"
)

type DiscoveryServer interface {
	init(discovery *Discovery, advertise *Advertise) error
	keep(second int)
	register() (interface{}, error)
	watch()
}

type Discovery struct {
	url    url.URL
	rawurl string
}

type Advertise struct {
	host   string
	ip     string
	origin string
}

const DISCOVERY_REGISTER_NODE = "register"
const DISCOVERY_KEEPALIVE = 60

var (
	server DiscoveryServer
)

// 开启服务
func StartServer(discovery *Discovery, advertise *Advertise) error {
	logger.Info("Start Server")
	logger.Debugf("Discovery info:%s", discovery)
	logger.Debugf("Advertise info :%s", advertise)
	server, err := InitServer(discovery, advertise)
	if err != nil {
		return err
	}
	server.keep(DISCOVERY_KEEPALIVE)
	return nil
}

func InitServer(discovery *Discovery, advertise *Advertise) (DiscoveryServer, error) {
	logger.Infof("DiscoveryServer is %s, init start", discovery.url.Scheme)
	switch discovery.url.Scheme {
	case "redis":
		server = new(RedisServer)
	}
	err := server.init(discovery, advertise)
	return server, err
}

func (discovery *Discovery) parse() error {
	if discovery.url.Scheme != "redis" {
		return fmt.Errorf("Discovery only support redis: %s", discovery.url.Scheme)
	}

	return nil
}

func (advertise *Advertise) parse() error {

	return nil
}
