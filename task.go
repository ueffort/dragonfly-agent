package main

import (
	"io"
	"os/exec"
)

type taskConfig struct {
	command  string
	argArray []string
}

type taskResult struct {
	state  int
	stdout io.Reader
	stderr io.Reader
}

func taskHandler(message interface) {
	
}

func taskRun(taskConfig *TaskConfig) (*TaskResult, error) {
	cmd := exec.Command(taskConfig.command, taskConfig.argArray...)
	taskResult := new(TaskResult)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	taskResult.stdout = stdout
	taskResult.stderr = stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return taskResult, nil
}
