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
			dpkMock := "2.20.11-0ubuntu82.4\tapport\tall\toptional\tutils\n"
			r := new(MockRunner)
			r.On("RunBytes", "apt-get", []string{"-s", "upgrade"}).Return(mockValue, nil)
			r.On("RunBytes", "dpkg-query", []string{"-W", "-f=" + dpkgQueryFormat, "apport"}).Return(dpkMock, nil)
			runner = r
			osPackages := packageManager.BuildPackageList()
			So(osPackages[0].Name, ShouldEqual, "apport")
			So(osPackages[0].Security, ShouldEqual, false)
			So(osPackages[0].Current_version, ShouldEqual, "2.20.11-0ubuntu82.4")
			So(osPackages[0].Candidate_version, ShouldEqual, "2.20.11-0ubuntu82.5")
			So(osPackages[0].Priority, ShouldEqual, "optional")
			So(osPackages[0].Section, ShouldEqual, "utils")
			So(osPackages[0].Source, ShouldEqual, "apport")
			So(osPackages[0].Arch, ShouldEqual, "all")
			So(osPackages[0].InstalledEVR, ShouldEqual, "2.20.11-0ubuntu82.4")
			r.Mock.AssertExpectations(t)
		})
	})
}

func TestPackageListSourceArchAndEVR(t *testing.T) {
	packageManager = DebianPackageManager{}
	Convey("Given pending updates whose binary and source names differ", t, func() {
		mockValue := `
Inst libssl3 [3.0.2-0ubuntu1.15] (3.0.2-0ubuntu1.18 Debian-Security:12/stable-security [amd64])
Inst libc6:i386 [2.35-0ubuntu3.6] (2.35-0ubuntu3.8 Ubuntu:22.04/jammy-updates [i386])
Inst tzdata [1:2024a-0ubuntu0.22.04] (1:2024b-0ubuntu0.22.04 Ubuntu:22.04/jammy-updates [all])
`
		r := new(MockRunner)
		r.On("RunBytes", "apt-get", []string{"-s", "upgrade"}).Return(mockValue, nil)
		r.On("RunBytes", "dpkg-query", []string{"-W", "-f=" + dpkgQueryFormat, "libssl3"}).
			Return("3.0.2-0ubuntu1.15\topenssl\tamd64\toptional\tlibs\n", nil)
		r.On("RunBytes", "dpkg-query", []string{"-W", "-f=" + dpkgQueryFormat, "libc6:i386"}).
			Return("2.35-0ubuntu3.6\tglibc\ti386\toptional\tlibs\n", nil)
		// older dpkg without source:Package support yields an empty source
		r.On("RunBytes", "dpkg-query", []string{"-W", "-f=" + dpkgQueryFormat, "tzdata"}).
			Return("1:2024a-0ubuntu0.22.04\t\tall\trequired\tlocalization\n", nil)
		runner = r

		pkgs := packageManager.BuildPackageList()
		So(len(pkgs), ShouldEqual, 3)

		So(pkgs[0].Name, ShouldEqual, "libssl3")
		So(pkgs[0].Source, ShouldEqual, "openssl")
		So(pkgs[0].Arch, ShouldEqual, "amd64")
		So(pkgs[0].InstalledEVR, ShouldEqual, "3.0.2-0ubuntu1.15")
		So(pkgs[0].Current_version, ShouldEqual, "3.0.2-0ubuntu1.15")
		So(pkgs[0].Security, ShouldBeTrue)

		So(pkgs[1].Name, ShouldEqual, "libc6:i386")
		So(pkgs[1].Source, ShouldEqual, "glibc")
		So(pkgs[1].Arch, ShouldEqual, "i386")

		So(pkgs[2].Source, ShouldEqual, "tzdata")
		So(pkgs[2].InstalledEVR, ShouldEqual, "1:2024a-0ubuntu0.22.04")
		So(pkgs[2].Current_version, ShouldEqual, "1:2024a-0ubuntu0.22.04")
		So(pkgs[2].Priority, ShouldEqual, "required")
		So(pkgs[2].Section, ShouldEqual, "localization")
		r.Mock.AssertExpectations(t)
	})
}

func TestParseDpkgQuery(t *testing.T) {
	Convey("Given dpkg-query output", t, func() {
		Convey("Multi-arch duplicates use the first line", func() {
			info := parseDpkgQuery("libc6", []byte("2.35-0ubuntu3.6\tglibc\tamd64\trequired\tlibs\n2.35-0ubuntu3.6\tglibc\ti386\trequired\tlibs\n"))
			So(info.Version, ShouldEqual, "2.35-0ubuntu3.6")
			So(info.Arch, ShouldEqual, "amd64")
		})
		Convey("Empty source falls back to the name without arch qualifier", func() {
			info := parseDpkgQuery("libfoo:i386", []byte("1.0-1\t\ti386\toptional\tlibs\n"))
			So(info.Source, ShouldEqual, "libfoo")
		})
		Convey("Short output does not panic", func() {
			info := parseDpkgQuery("foo", []byte("1.0-1"))
			So(info.Version, ShouldEqual, "1.0-1")
			So(info.Source, ShouldEqual, "foo")
			So(info.Arch, ShouldEqual, "")
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
