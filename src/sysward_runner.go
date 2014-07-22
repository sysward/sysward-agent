package main

import (
	"os"
	"os/exec"
)

type Runner interface {
	Run(string, ...string) (string, error)
}

type SyswardRunner struct{}

func (r SyswardRunner) Run(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	path := os.Getenv("PATH")
	cmd.Env = []string{"DEBIAN_FRONTEND=noninteractive", "PATH=" + path}
	out, err := cmd.CombinedOutput()
	return string(out), err
}
