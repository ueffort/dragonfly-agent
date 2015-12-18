package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	flag "github.com/ueffort/goutils/mflag"
)

var (
	version = "0.1.1"
)

func main() {

	flHelp := flag.Bool([]string{"h", "-help"}, false, "Print usage")
	flVersion := flag.Bool([]string{"v", "-version"}, false, "Print version information and quit")

	flag.Usage = func() {
		fmt.Fprint(os.Stdout, "Usage: agent [OPTIONS] \n       agent [ -h | --help | -v | --version ]\n\n")
		fmt.Fprint(os.Stdout, "Dragonfly Agent component.\n\nOptions:")

		flag.CommandLine.SetOutput(os.Stdout)
		flag.PrintDefaults()
	}

	flag.Merge(flag.CommandLine, envFlags.FlagSet, commonFlags.FlagSet)
	flag.Parse()

	if *flVersion {
		showVersion()
		return
	}

	if *flHelp {
		flag.Usage()
		return
	}

	err := envFlags.PostParse()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}
	err = commonFlags.PostParse()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}
	if !enoughFlag {
		flag.Usage()
		return
	}

	err = Run()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	go SignalHandle()
	time.Sleep(time.Hour)
}

func showVersion() {
	fmt.Printf("agent version %s\n", version)
}
func SignalHandle() {
	ss := signalSetNew()
	handler := func(s os.Signal, arg interface{}) {
		fmt.Printf("handle signal: %v\n", s)
	}

	ss.register(syscall.SIGINT, handler)
	ss.register(syscall.SIGUSR1, handler)
	ss.register(syscall.SIGUSR2, handler)

	for {
		c := make(chan os.Signal)
		var sigs []os.Signal
		for sig := range ss.m {
			sigs = append(sigs, sig)
		}
		signal.Notify(c)
		sig := <-c

		err := ss.handle(sig, nil)
		if err != nil {
			fmt.Printf("unknown signal received: %v\n", sig)
			os.Exit(1)
		}
	}
}

type signalHandler func(s os.Signal, arg interface{})

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
		set.m[sig](sig, arg)
		return nil
	} else {
		return fmt.Errorf("No handler available for signal %v", sig)
	}

	panic("won't reach here")
}
