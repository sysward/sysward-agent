package main

import (
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFileAppends(t *testing.T) {
	Convey("Appending to a file", t, func() {
		f := SyswardFileWriter{}
		file, _ := os.Create("/tmp/sysward_writer_test")
		file.Close()
		f.AppendToFile("/tmp/sysward_writer_test", "testing")
		r := SyswardFileReader{}
		tmpFile, _ := r.ReadFile("/tmp/sysward_writer_test")
		So(strings.Contains(string(tmpFile), "testing"), ShouldBeTrue)
	})
}
