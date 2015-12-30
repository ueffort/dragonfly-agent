package main

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestSignal(t *testing.T) {
	accept := make(chan bool)
	closeHandler := func(s os.Signal, arg interface{}) error {
		accept <- true
		return nil
	}
	go SignalHandle(closeHandler)
	<-time.After(time.Second * time.Duration(2))
	syscall.Kill(os.Getpid(), syscall.SIGINT)
	select {
	case <-accept:
		t.Log("close success")
	case <-time.After(time.Second * time.Duration(2)):
		t.Error("close error")
		return
	}
}
