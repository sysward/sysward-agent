package main

import (
	"fmt"
	"github.com/sysward/sysward-agent/logging"
	"os"
	"os/exec"
)

type Runner interface {
	Run(string, ...string) (string, error)
}

type SyswardRunner struct{}

func (r SyswardRunner) Run(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	out, err := cmd.CombinedOutput()
	if os.Getenv("DEBUG") == "true" {
		logging.LogMsg(fmt.Sprintf("Command: %s %#v", command, args))
	}
	return string(out), err
}
