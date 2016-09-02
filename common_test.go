package main

import (
	"testing"
	"time"
)

var (
	discoverRedis *Discovery
)

func init() {
	s := "redis://root@127.0.0.1:6379/test/master"
	discoverRedis, _ = parseDiscovery(&s)
}

func TestRedis(t *testing.T) {
	var server DiscoveryServer
	start := make(chan bool)
	stop := make(chan bool)
	accept := make(chan []byte)
	server = new(Redis)
	err := server.init(discoverRedis)
	if err != nil {
		t.Error(err)
		return
	}
	handle := func(message []byte) {
		accept <- message
	}
	cancelWatch, err := server.watch("master/test", handle, start, stop)
	if err != nil {
		t.Log(err)
		return
	}
	select {
	case <-start:
		t.Log("start success")
	case <-time.After(time.Second * time.Duration(5)):
		t.Error("start error")
		return
	}
	server.notice("test", "test")
	select {
	case <-accept:
		t.Log("notice success")
	case <-time.After(time.Second * time.Duration(5)):
		t.Error("notice error")
		return
	}
	cancelWatch()
	server.notice("test", "test")
	select {
	case <-accept:
		t.Error("cancel error")
		return
	case <-time.After(time.Second * time.Duration(2)):
		t.Log("cancel success")
	}
}
