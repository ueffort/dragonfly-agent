package main

import (
	"testing"
)

func TestCli(t *testing.T) {
	envFlags.Debug = true
	envFlags.LogLevel = "debug"
	err := envFlags.PostParse()
	if err != nil {
		t.Error(err)
	}
	commonFlags.ConfigFileName = "config.json"
	err = commonFlags.PostParse()
	if err != nil {
		t.Error(err)
	}

	discovery, err := parseDiscovery(&commonFlags.Discovery)
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("discovery :%s", discovery)
	}
	advertise, err := parseAdvertise(&commonFlags.Advertise)
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("advertise :%s", advertise)
	}

}
