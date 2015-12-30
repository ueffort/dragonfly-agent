package main

import (
	"testing"

	"github.com/golang/protobuf/proto"
)

func TestTask(t *testing.T) {
	target := DISCOVERY_EXCEPTION
	newTask := new(Task)
	id := "123"
	newTask.Id = &id
	channel := "master"
	newTask.Channel = &channel
	command := "pwd"
	newTask.Command = &command
	message, err := proto.Marshal(newTask)
	message_r, err := protoHandler(&target, message)
	if err != nil {
		t.Error(err)
		return
	} else {
		t.Log(message_r)
	}
}
