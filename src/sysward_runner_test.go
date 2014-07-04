package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestExecutingARealCommand(t *testing.T) {
	runner = SyswardRunner{}
	Convey("The command exits properly", t, func() {
		out, err := runner.Run("echo", "hello world")
		So(out, ShouldEqual, "hello world\n")
		So(err, ShouldBeNil)
	})
	Convey("The command doesnt exit properly", t, func() {
		_, err := runner.Run("grep", "foo", "/tmp/fakefile")
		So(err, ShouldNotBeNil)
	})
}
