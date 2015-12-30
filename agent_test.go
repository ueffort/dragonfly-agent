package main

import (
	"os"
	"testing"

	flag "github.com/ueffort/goutils/mflag"
)

func TestFlag(t *testing.T) {
	os.Args = make([]string, 2)
	code, needExit := parseFlag()
	if needExit {
		t.Errorf("parse error: code %s", code)
	} else {
		t.Log("parse success")
	}
	showVersion()
	flag.Usage()
}
