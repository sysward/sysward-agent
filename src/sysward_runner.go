package main

import (
	"os/exec"
)

type Runner interface {
	Run(string, ...string) ([]byte, error)
}

type SyswardRunner struct{}

func (r SyswardRunner) Run(command string, args ...string) ([]byte, error) {
	out, err := exec.Command(command, args...).CombinedOutput()
	return out, err
}
