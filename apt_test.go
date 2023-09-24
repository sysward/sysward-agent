package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPackagesThatNeedUpdates(t *testing.T) {
	packageManager = DebianPackageManager{}
	Convey("Given pending updates", t, func() {

		Convey("There should be a list of packages available for update", func() {
			mockValue := `
Inst apport [2.20.11-0ubuntu82.4] (2.20.11-0ubuntu82.5 Ubuntu:22.04/jammy-updates [all])
`
			dpkMock := `
Package: apport
Status: install ok installed
Priority: optional
Section: utils
Installed-Size: 812
Maintainer: Ubuntu Developers <ubuntu-devel-discuss@lists.ubuntu.com>
Architecture: all
Version: 2.20.11-0ubuntu82.4
Replaces: core-dump-handler, python-apport (<< 2.2-0ubuntu1)
Provides: core-dump-handler
Depends: python3, python3-apport (>= 2.20.11-0ubuntu82.4), lsb-base (>= 3.0-6), python3-gi, gir1.2-glib-2.0 (>= 1.29.17)
Recommends: apport-symptoms, python3-systemd
Suggests: apport-gtk | apport-kde, policykit-1
Breaks: python-apport (<< 2.2-0ubuntu1)
Conflicts: core-dump-handler
Conffiles:
 /etc/apport/blacklist.d/README.blacklist c2ed1eb9a17ec2550747b4960cf4b73c
 /etc/apport/blacklist.d/apport 44503501302b80099552bac0204a45c1
 /etc/apport/crashdb.conf 4202dae3eccfa5bbb33a0a9acfcd3724
 /etc/bash_completion.d/apport_completion dfe766d9328bb5c895038b44185133f9
 /etc/cron.daily/apport df5d3bc9ab3a67b58156376318077304
 /etc/default/apport 3446c6cac185f44237f59786e006ebe4
 /etc/init.d/apport 3d51dc9135014bb49b4a19ff8dab61f1
 /etc/logrotate.d/apport fa54dab59ef899b48d5455c976008df4
Description: automatically generate crash reports for debugging
 apport automatically collects data from crashed processes and
 compiles a problem report in /var/crash/. This utilizes the crashdump
 helper hook provided by the Ubuntu kernel.
 .
 This package also provides a command line frontend for browsing and
 handling the crash reports. For desktops, you should consider
 installing the GTK+ or Qt user interface (apport-gtk or apport-kde).
Homepage: https://wiki.ubuntu.com/Apport
`
			r := new(MockRunner)
			r.On("RunBytes", "apt-get", []string{"-s", "upgrade"}).Return(mockValue, nil)
			r.On("RunBytes", "dpkg", []string{"-s", "apport"}).Return(dpkMock, nil)
			runner = r
			osPackages := packageManager.BuildPackageList()
			So(osPackages[0].Name, ShouldEqual, "apport")
			So(osPackages[0].Security, ShouldEqual, false)
			r.Mock.AssertExpectations(t)
		})
	})
}

func TestChangeLog(t *testing.T) {
	packageManager = DebianPackageManager{}
	Convey("Changelong gets 64bit encoded", t, func() {
		r := new(MockRunner)
		r.On("Run", "apt-get", []string{"changelog", "apt"}).Return("foobar", nil)
		runner = r
		So(packageManager.GetChangelog("apt"), ShouldEqual, "Zm9vYmFy")
		r.Mock.AssertExpectations(t)
	})
}

func TestPackageUpdates(t *testing.T) {
	packageManager = DebianPackageManager{}
	Convey("Given a package name", t, func() {

		Convey("The package should be upgraded", func() {
			r := new(MockRunner)
			r.On("Run", "apt-get", []string{
				"install",
				"-y",
				"-o",
				fmt.Sprintf("Dpkg::Options::=--force-confdef"),
				"-o",
				fmt.Sprintf("Dpkg::Options::=--force-confold"),
				"apt"}).Return("", nil)
			runner = r
			err := packageManager.UpdatePackage("apt")
			So(err, ShouldBeNil)
			r.Mock.AssertExpectations(t)
		})

		Convey("The package should not upgrade if held", func() {
			r := new(MockRunner)
			r.On("Run", "apt-get", []string{
				"install",
				"-y",
				"-o",
				fmt.Sprintf("Dpkg::Options::=--force-confdef"),
				"-o",
				fmt.Sprintf("Dpkg::Options::=--force-confold"),
				"apt"}).Return("", errors.New("fail"))
			runner = r
			err := packageManager.UpdatePackage("apt")
			So(err, ShouldNotBeNil)
			r.Mock.AssertExpectations(t)
		})
	})

}

func TestPackageHolding(t *testing.T) {
	packageManager = DebianPackageManager{}
	Convey("Given holding a package", t, func() {

		Convey("The package should be held", func() {
			r := new(MockRunner)
			r.On("Run", "apt-mark", []string{"hold", "apt"}).Return("", nil)
			runner = r
			err := packageManager.HoldPackage("apt")
			So(err, ShouldBeNil)
			r.Mock.AssertExpectations(t)
		})
	})

	Convey("Given unholding a package", t, func() {

		Convey("The package should be unheld", func() {
			r := new(MockRunner)
			r.On("Run", "apt-mark", []string{"unhold", "apt"}).Return("", nil)
			runner = r
			err := packageManager.UnholdPackage("apt")
			So(err, ShouldBeNil)
			r.Mock.AssertExpectations(t)
		})
	})

}

func TestSourceList(t *testing.T) {
	packageManager = DebianPackageManager{}
	Convey("Given /etc/apt/sources.list exists", t, func() {

		Convey("There should be a source list", func() {
			packageList := []string{"deb http://us.archive.ubuntu.com/ubuntu/ trusty main restricted",
				"deb-src http://us.archive.ubuntu.com/ubuntu/ trusty main restricted"}
			r := new(MockRunner)
			r.On("Run", "grep", []string{"-h", "^deb", "/etc/apt/sources.list", "/etc/apt/sources.list.d/*"}).Return(strings.Join(packageList, "\n"), nil)
			runner = r
			sourcesList := packageManager.GetSourcesList()
			src_one := sourcesList[0]
			src_two := sourcesList[1]
			So(src_one.Url, ShouldEqual, "http://us.archive.ubuntu.com/ubuntu/")
			So(src_one.Src, ShouldBeFalse)
			So(src_two.Src, ShouldBeTrue)
			r.Mock.AssertExpectations(t)
		})
	})

}

func TestInstalledPackages(t *testing.T) {
	packageManager = DebianPackageManager{}
	Convey("Given I want to view all installed packages", t, func() {
		Convey("It returns a list of all installed packages", func() {
			r := new(MockRunner)
			r.On("Run", "dpkg", []string{"--get-selections"}).Return("apt\u0009installed", nil)
			runner = r
			packages := packageManager.BuildInstalledPackageList()
			So(packages[0], ShouldEqual, "apt")
			So(len(packages), ShouldEqual, 1)
			r.Mock.AssertExpectations(t)
		})
	})

}

func TestUpdatingThePackageList(t *testing.T) {
	packageManager = DebianPackageManager{}
	Convey("Given I want to have the latest source list", t, func() {

		Convey("apt-update gets run", func() {
			r := new(MockRunner)
			r.On("Run", "apt-get", []string{"update"}).Return("", nil)
			runner = r
			err := packageManager.UpdatePackageLists()
			So(err, ShouldBeNil)
			r.Mock.AssertExpectations(t)
		})
	})

}

//func TestUpdateCounts(t *testing.T) {
//	packageManager = DebianPackageManager{}
//	Convey("Given there are security and regular updates", t, func() {
//
//		Convey("The number of security and regular updates is > 0", func() {
//			//r := new(MockRunner)
//			//r.On("Run", "/usr/lib/update-notifier/apt-check", []string{}).Return("1;2", nil)
//			//runner = r
//			updates := packageManager.UpdateCounts()
//			So(updates.Regular, ShouldEqual, 1)
//			So(updates.Security, ShouldEqual, 2)
//			//r.Mock.AssertExpectations(t)
//		})
//
//		Convey("There are no security updates", func() {
//			r := new(MockRunner)
//			r.On("Run", "/usr/lib/update-notifier/apt-check", []string{}).Return("2;0", nil)
//			runner = r
//			updates := packageManager.UpdateCounts()
//			So(updates.Regular, ShouldEqual, 2)
//			So(updates.Security, ShouldEqual, 0)
//			r.Mock.AssertExpectations(t)
//		})
//
//	})
//}
