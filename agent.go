package main

import (
	"fmt"
	"os"

	flag "github.com/ueffort/goutils/mflag"
)

var (
	version = "0.1.1"
)

func main() {
	code, needExit := parseFlag()
	if needExit {
		os.Exit(code)
	}

	err := Run()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

func parseFlag() (int, bool) {

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
		return 0, true
	}

	if *flHelp {
		flag.Usage()
		return 0, true
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
		return 1, true
	}
	return 0, false
}

func showVersion() {
	fmt.Printf("agent version %s\n", version)
}
