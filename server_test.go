package main

import (
	"testing"
)

func TestDiscovery(t *testing.T) {
	discoveryStr := "redis://root@127.0.0.1:6379/dragonfly/master"
	discovery := new(Discovery)
	err := discovery.parse(discoveryStr)
	if err != nil {
		t.Error(err)
	}
}

func TestAdvertise(t *testing.T) {
	advertiseStr := "127.0.0.1"
	advertise := new(Advertise)
	err := advertise.parse(advertiseStr)
	if err != nil {
		t.Error(err)
	}
}

func TestServer(t *testing.T) {
	sm := new(ServerManage)
	discoveryStr := "redis://root@127.0.0.1:6379/dragonfly/master"
	discovery := new(Discovery)
	discovery.parse(discoveryStr)
	advertiseStr := "127.0.0.1"
	advertise := new(Advertise)
	advertise.parse(advertiseStr)
	server, err := discovery.init()
	if err != nil {
		t.Error(err)
		return
	}
	sm.server = server
	sm.quit = make(chan bool)
	sm.clear = make(chan bool)
	sm.run(advertise)
	sm.stop()
}
