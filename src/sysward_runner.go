package main

import (
	"fmt"
	"os"
	"os/exec"
	"sysward_agent/src/logging"
)

type Runner interface {
	Run(string, ...string) (string, error)
}

type SyswardRunner struct{}

func (r SyswardRunner) Run(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	//path := os.Getenv("PATH")
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	//cmd.Env = []string{"DEBIAN_FRONTEND=noninteractive", "PATH=" + path}
	out, err := cmd.CombinedOutput()
	if os.Getenv("DEBUG") == "true" {
		logging.LogMsg(fmt.Sprintf("Command: %s %#v", command, args))
	}
	return string(out), err
}
