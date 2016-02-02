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
	target, message_r, err := protoHandler(message)
	if err != nil {
		t.Error(err)
		return
	} else if target != channel {
		t.Error(target)
	} else {
		t.Log(target)
		t.Log(message_r)
	}
}
