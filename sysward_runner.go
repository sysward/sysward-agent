package main

import (
	"fmt"
	"github.com/sysward/sysward-agent/logging"
	"os"
	"os/exec"
)

type Runner interface {
	Run(string, ...string) (string, error)
	RunBytes(string, ...string) ([]byte, error)
}

type SyswardRunner struct{}

func (r SyswardRunner) RunBytes(command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	out, err := cmd.CombinedOutput()
	if os.Getenv("DEBUG") == "true" {
		logging.LogMsg(fmt.Sprintf("Command: %s %#v", command, args))
	}
	return out, err
}

func (r SyswardRunner) Run(command string, args ...string) (string, error) {
	out, err := r.RunBytes(command, args...)
	return string(out), err
}
