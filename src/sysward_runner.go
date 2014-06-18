package main

import (
	"os/exec"
)

type SyswardRunner struct{}

func (r SyswardRunner) Run(command string, args ...string) ([]byte, error) {
	out, err := exec.Command(command, args...).CombinedOutput()
	return out, err
}
