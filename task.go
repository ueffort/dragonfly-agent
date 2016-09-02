package main

import (
	"io/ioutil"
	"os/exec"

	"github.com/golang/protobuf/proto"
)

func protoHandler(message []byte) (string, []byte, error) {
	newTask := new(Task)
	newResult := new(Result)
	err := proto.Unmarshal(message, newTask)
	if err != nil {
		newResult.State = TaskState_EXCEPTION.Enum()
		return "", nil, err
	}
	newResult.Id = newTask.Id
	logger.Debugf("Task message :%s", newTask)
	err = newTask.run(newResult)
	if err != nil {
		logger.Error(err)
		newResult.State = TaskState_FAIL.Enum()
	} else {
		newResult.State = TaskState_SUCCESS.Enum()
	}
	message, err = proto.Marshal(newResult)
	logger.Debugf("Result message :%s", newResult)
	return *newTask.Channel, message, err
}

func (task *Task) run(result *Result) error {
	s := ""
	result.Out = &s
	result.Err = &s
	cmd := exec.Command(*task.Command)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	outb, o := ioutil.ReadAll(stdout)
	errb, f := ioutil.ReadAll(stderr)
	if err := cmd.Wait(); err != nil {
		return err
	}
	if o != nil {
		return o
	}
	if f != nil {
		return f
	}
	result.Out = task.io(&outb)
	result.Err = task.io(&errb)
	return nil
}

func (task *Task) io(b *[]byte) *string {
	s := string(*b)
	b = nil
	return &s
}
