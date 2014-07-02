package main

import (
	"os/exec"
)

type Runner interface {
	Run(string, ...string) (string, error)
}

type SyswardRunner struct{}

func (r SyswardRunner) Run(command string, args ...string) (string, error) {
	out, err := exec.Command(command, args...).CombinedOutput()
	return string(out), err
}
