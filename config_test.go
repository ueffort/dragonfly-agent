package main

import (
	"testing"
)

func TestConfig(t *testing.T) {
	config, err := Load("config.json")
	if err != nil {
		t.Errorf("config error:%s", err)
	} else {
		t.Logf("config info:%s", config)
	}
}
