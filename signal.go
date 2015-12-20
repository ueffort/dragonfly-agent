package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type signalHandler func(s os.Signal, arg interface{}) error

type signalSet struct {
	m map[os.Signal]signalHandler
}

func signalSetNew() *signalSet {
	ss := new(signalSet)
	ss.m = make(map[os.Signal]signalHandler)
	return ss
}

func (set *signalSet) register(s os.Signal, handler signalHandler) {
	if _, found := set.m[s]; !found {
		set.m[s] = handler
	}
}

func (set *signalSet) handle(sig os.Signal, arg interface{}) (err error) {
	if _, found := set.m[sig]; found {
		return set.m[sig](sig, arg)
	} else {
		return fmt.Errorf("unknown signal received: %v\n", sig)
	}

	panic("won't reach here")
}

func SignalHandle() {
	ss := signalSetNew()
	closeHandler := func(s os.Signal, arg interface{}) error {
		return StopServer()
	}

	ss.register(syscall.SIGINT, closeHandler)
	ss.register(syscall.SIGHUP, closeHandler)
	ss.register(syscall.SIGQUIT, closeHandler)
	ss.register(syscall.SIGTERM, closeHandler)

	for {
		c := make(chan os.Signal)
		signal.Notify(c)
		sig := <-c
		ss.handle(sig, nil)
	}
}
