package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewAgent(t *testing.T) {
	Convey("Setting up a new agent", t, func() {
		agent := NewAgent()
		So(agent.runner, ShouldHaveSameTypeAs, SyswardRunner{})
		So(agent.fileReader, ShouldHaveSameTypeAs, SyswardFileReader{})
		So(agent.packageManager, ShouldHaveSameTypeAs, DebianPackageManager{})
	})
}

func TestAgentStartup(t *testing.T) {
	Convey("Agent startup should verify root and check pre-req packages", t, func() {
		r := new(MockRunner)
		f := new(MockReader)
		r.On("Run", "whoami", []string{}).Return("root", nil)
		runner = r
		f.On("FileExists", "/usr/lib/update-notifier/apt-check").Return(true)
		file_reader = f
		agent := Agent{}
		agent.Startup()
		f.Mock.AssertExpectations(t)
		r.Mock.AssertExpectations(t)
	})
}
