package main

import (
	"fmt"
	"net/url"
	"time"
)

type messageHandler func(message []byte)
type cancelWatch func()

type DiscoveryServer interface {
	init(discovery *Discovery) error
	// 内部异步监听,并将取消监听操作回调
	watch(advertise string, handler messageHandler, watching chan<- bool, watchoff chan<- bool) (cancelWatch, error)
	notice(target string, message interface{}) (interface{}, error)
}

type ServerManage struct {
	server DiscoveryServer
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

// 异常任务节点
const DISCOVERY_EXCEPTION = "exception"

// 心跳时间间隔
const DISCOVERY_KEEPALIVE_INTERVAL = 60

//断开等待时间
const DISCOVERY_BREAK_WAITTIME = 60

var (
	sm    *ServerManage
	quit  chan bool
	clear chan bool
)

func init() {
	sm = new(ServerManage)
	quit = make(chan bool)
	clear = make(chan bool)
}

// 开启服务
func StartServer(discovery *Discovery, advertise *Advertise) error {
	logger.Info("Start Server...")
	logger.Debugf("Discovery info:%s", discovery)
	logger.Debugf("Advertise info :%s", advertise)
	server, err := discovery.init()
	if err != nil {
		return err
	}
	sm.server = server
	go sm.run(advertise)
	return nil
}

//结束服务
func StopServer(exit chan<- bool) error {
	if sm == nil {
		return fmt.Errorf("Server is not init")
	}
	logger.Info("Stop Server...")
	quit <- true
	<-clear
	logger.Info("Clear over, can exit")
	exit <- true
	return nil
}

//实现server的keepalive
func (sm *ServerManage) keep(second int, advertise *Advertise, working <-chan bool, workoff <-chan bool, unkeep <-chan bool) {
	register := func() {
		replay, err := sm.server.notice(DISCOVERY_REGISTER, advertise.origin)
		logger.Infof(
			"Register advertise %s, replay->%s, err->%s, wait %s second",
			advertise.origin,
			replay,
			err,
			second)
	}
	unregister := func() {
		replay, err := sm.server.notice(DISCOVERY_UNREGISTER, advertise.origin)
		logger.Infof(
			"UnRegister advertise %s, replay->%s, err->%s, wait work",
			advertise.origin,
			replay,
			err,
			second)
	}
	for {
		select {
		case <-working:
			register()
		case <-workoff:
			unregister()
		case <-unkeep:
			unregister()
			return
		case <-time.After(time.Second * time.Duration(second)):
			register()
		}
	}
}

func (sm *ServerManage) work(second int, advertise *Advertise, working chan<- bool, workoff chan<- bool, unwork <-chan bool) {
	except := make(chan bool)
	var cancel cancelWatch
	var err error
	on := func() {
		cancel, err = sm.server.watch(advertise.origin, sm.task, working, except)
		logger.Infof("Watch advertise %s, err->%s", advertise.origin, err)
	}
	off := func() {
		if cancel != nil {
			cancel()
		}
		logger.Infof("Cancel Watch advertise %s", advertise.origin)
	}
	on()
	for {
		select {
		case <-unwork:
			off()
			return
		case <-except:
			cancel = nil
			logger.Infof("Watch is break, wait %s second reconnect", second)
			workoff <- true
			select {
			case <-unwork:
				off()
			case <-time.After(time.Second * time.Duration(second)):
				on()
			}
		}
	}
}

//处理task
func (sm *ServerManage) task(message []byte) {
	go func() {
		target := DISCOVERY_EXCEPTION
		message_r, err := protoHandler(&target, message)
		logger.Debug("Task end, target->%s, err->%s", target, err)

		result, err := sm.server.notice(target, message_r)
		logger.Debug("Send result, result->%s, err->%s", result, err)
	}()
}

//server的运行操作
func (sm *ServerManage) run(advertise *Advertise) {
	working := make(chan bool)
	workoff := make(chan bool)
	unwork := make(chan bool)
	unkeep := make(chan bool)
	stopwork := make(chan bool)
	stopkeep := make(chan bool)
	work := func(on chan<- bool, off chan<- bool, stop <-chan bool, end chan<- bool) {
		logger.Infof("Server work advertise:%s", advertise)
		sm.work(DISCOVERY_BREAK_WAITTIME, advertise, on, off, stop)
		logger.Infoln("Server work stoped")
		end <- true
	}
	keep := func(start <-chan bool, pause <-chan bool, stop <-chan bool, end chan<- bool) {
		logger.Infof("KeepAlive interval %s second", DISCOVERY_KEEPALIVE_INTERVAL)
		sm.keep(DISCOVERY_KEEPALIVE_INTERVAL, advertise, start, pause, stop)
		logger.Infoln("KeepAlive stoped")
		end <- true
	}
	go work(working, workoff, unwork, stopwork)
	go keep(working, workoff, unkeep, stopkeep)
	for {
		select {
		case <-quit:
			unwork <- true
			unkeep <- true
			<-stopwork
			<-stopkeep
			clear <- true
			return
		}
	}
}

//根据discovery参数生成server
func (discovery *Discovery) init() (DiscoveryServer, error) {
	var server DiscoveryServer
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
	advertise.tag = make([]string, 0)
	return nil
}
