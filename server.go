package main

import (
	"fmt"
	"net/url"
)

type messageHandler func(message interface{}) error

type DiscoveryServer interface {
	init(discovery *Discovery) error
	watch(advertise string, handler messageHandler, watching chan<- bool, unwartch <-chan bool)
	notice(target string, message interface{}) (interface{}, error)
}

type Discovery struct {
	url    url.URL
	rawurl string
}

type Advertise struct {
	tag    []string
	origin string
}

// 注册agent节点
const DISCOVERY_REGISTER = "register"

// 注销agent节点
const DISCOVERY_UNREGISTER = "unregister"

// 心跳时间间隔
const DISCOVERY_KEEPALIVE_INTERVAL = 60

//断开等待时间
const DISCOVERY_BREAK_WAITTIME = 60

var (
	server DiscoveryServer
	quit   chan bool
	clear  chan bool
)

// 开启服务
func StartServer(discovery *Discovery, advertise *Advertise) error {
	logger.Info("Start Server...")
	logger.Debugf("Discovery info:%s", discovery)
	logger.Debugf("Advertise info :%s", advertise)
	server, err := discovery.init()
	if err != nil {
		return err
	}
	quit = make(chan bool)
	clear = make(chan bool)
	go server.run(advertise)
	return nil
}

//结束服务
func StopServer() error {
	if server == nil {
		return fmt.Errorf("Server is not init")
	}
	logger.Info("Stop Server...")
	quit <- true
	<-clear
	exit <- true
}

//实现server的keepalive
func (server DiscoveryServer) keep(second int, advertise *Advertise, watching <-chan bool, stopwatch <-chan bool, unkeep <-chan bool) {
	register = func() {
		replay, err := server.notice(DISCOVERY_REGISTER, advertise.origin)
		logger.Infof(
			"Register advertise %s, replay->%s, err->%s, wait %s second",
			advertise.origin,
			replay,
			err,
			second)
	}
	unregister = func() {
		replay, err := server.notice(DISCOVERY_UNREGISTER, advertise.origin)
		logger.Infof(
			"UNRegister advertise %s, replay->%s, err->%s, wait %s second",
			advertise.origin,
			replay,
			err,
			second)
	}
	for {
		select {
		case <-watching:
			register()
		case <-stopwatch:
			unregister()
		case <-unkeep:
			unregister()
			return
		case <-time.After(time.Second * time.Duration(second)):
			register()
		}
	}
}

//处理server的运行操作
func (server DiscoveryServer) run(advertise *Advertise) {
	unwatch := make(chan bool)
	watching := make(chan bool)
	unkeep := make(chan bool)
	stopwatch := make(chan bool)
	stopkeey := make(chan bool)
	watch = func(start chan<- bool, stop <-chan bool, end chan<- bool) {
		logger.Infof("Watching, wait message", DISCOVERY_KEEPALIVE_INTERVAL)
		server.watch(advertise.origin, taskHandler, start, stop)
		end <- true
	}
	keep = func(start <-chan bool, wait <-chan bool, stop <-chan bool, end chan<- bool) {
		logger.Infof("KeepAlive interval %s second", DISCOVERY_KEEPALIVE_INTERVAL)
		server.keep(DISCOVERY_KEEPALIVE_INTERVAL, advertise, start, wait, stop)
		end <- true
	}
	go watch(watching, unwatch, stopwatch)
	go keep(watching, stopwatch, unkeep, stopkeep)
	for {
		switch {
		case <-quit:
			unwatch <- true
			unkeep <- true
			<-stopwatch
			<-stopkeep
			clear <- true
			return
		case <-stopwatch:
			logger.Infof("Watch is break, wait %s second reconnect", DISCOVERY_BREAK_WAITTIME)
			<-time.After(time.Second * time.Duration(second))
			go watch(watching, unwatch, stopwatch)
		}
	}
}

//根据discovery参数生成server
func (discovery *Discovery) init() (DiscoveryServer, error) {
	logger.Infof("DiscoveryServer is %s, init start", discovery.url.Scheme)
	switch discovery.url.Scheme {
	case "redis":
		server = new(Redis)
	}
	err := server.init(discovery)
	return server, err
}

//根据discovery配置格式化参数信息
func (discovery *Discovery) parse(discoveryStr string) error {
	u, err := url.Parse(discoveryStr)
	if err != nil {
		return err
	}
	if u.Scheme != "redis" {
		return fmt.Errorf("Discovery only support redis: %s", discovery.url.Scheme)
	}

	discovery.url = *u
	discovery.rawurl = discoveryStr
	return nil
}

//根据advertise配置格式化参数信息
func (advertise *Advertise) parse(advertiseStr string) error {
	advertise.origin = advertiseStr
	advertise.tag = make([]string)
	return nil
}
