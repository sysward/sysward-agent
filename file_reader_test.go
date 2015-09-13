package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFileExists(t *testing.T) {
	Convey("A file should exist", t, func() {
		f := SyswardFileReader{}
		So(f.FileExists("config.json"), ShouldBeTrue)
	})

	Convey("A file doesnt exist", t, func() {
		f := SyswardFileReader{}
		So(f.FileExists("/tmp/fakefile"), ShouldBeFalse)
	})
}
