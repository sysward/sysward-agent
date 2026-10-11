package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestZypperBuildPackageList(t *testing.T) {
	Convey("Given zypper list-updates output with an Arch column", t, func() {
		r := new(MockRunner)
		r.On("Run", "zypper", []string{"list-updates"}).Return(`Loading repository data...
Reading installed packages...
S | Repository     | Name    | Current Version    | Available Version  | Arch
--+----------------+---------+--------------------+--------------------+-------
v | Update repo    | openssl | 1.1.1l-150400.7.25 | 1.1.1l-150400.7.28 | x86_64
v | Update repo    | tzdata  | 2024a-150000.3.103 | 2024b-150000.3.106 | noarch
`, nil)
		r.On("Run", "zypper", []string{"list-patches"}).Return(`Repository | Name | Category | Severity | Interactive | Status | Summary
Update | openSUSE-SLE-15.4-2024-1 | security | important | --- | needed | Security update for openssl
`, nil)
		runner = r

		pkgs := ZypperPackageManager{}.BuildPackageList()
		So(len(pkgs), ShouldEqual, 2)
		So(pkgs[0].Name, ShouldEqual, "openssl")
		So(pkgs[0].Current_version, ShouldEqual, "1.1.1l-150400.7.25")
		So(pkgs[0].Candidate_version, ShouldEqual, "1.1.1l-150400.7.28")
		So(pkgs[0].InstalledEVR, ShouldEqual, "1.1.1l-150400.7.25")
		So(pkgs[0].Arch, ShouldEqual, "x86_64")
		So(pkgs[0].Source, ShouldEqual, "")
		So(pkgs[1].Arch, ShouldEqual, "noarch")
		r.Mock.AssertExpectations(t)
	})

	Convey("Given zypper list-updates output without an Arch column", t, func() {
		r := new(MockRunner)
		r.On("Run", "zypper", []string{"list-updates"}).Return("v | Update repo | bash | 4.4-1 | 4.4-2\n", nil)
		r.On("Run", "zypper", []string{"list-patches"}).Return("", nil)
		runner = r

		pkgs := ZypperPackageManager{}.BuildPackageList()
		So(len(pkgs), ShouldEqual, 1)
		So(pkgs[0].InstalledEVR, ShouldEqual, "4.4-1")
		So(pkgs[0].Arch, ShouldEqual, "")
	})
}
