package main

import (
	"errors"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPackagesThatNeedUpdates(t *testing.T) {
	Convey("Given pending updates", t, func() {

		Convey("There should be a list of packages available for update", func() {
			mockValue := `[{"name": "apt", "section": "admin", "priority": "important", "current_version": "1.0.1ubuntu2", "security": true, "candidate_version": "1.0.1ubuntu2.1"}]`
			r := new(MockRunner)
			r.On("Run", "python", []string{"trex.py"}).Return(mockValue, nil)
			runner = r
			osPackages := buildPackageList()
			So(osPackages[0].Name, ShouldEqual, "apt")
			So(osPackages[0].Security, ShouldEqual, true)
			r.Mock.AssertExpectations(t)
		})
	})

}

func TestPackageUpdates(t *testing.T) {
	Convey("Given a package name", t, func() {

		Convey("The package should be upgraded", func() {
			r := new(MockRunner)
			r.On("Run", "apt-get", []string{"install", "-y", "apt"}).Return("", nil)
			runner = r
			err := updatePackage("apt")
			So(err, ShouldBeNil)
			r.Mock.AssertExpectations(t)
		})

		Convey("The package should not upgrade if held", func() {
			r := new(MockRunner)
			r.On("Run", "apt-get", []string{"install", "-y", "apt"}).Return("", errors.New("fail"))
			runner = r
			err := updatePackage("apt")
			So(err, ShouldNotBeNil)
			r.Mock.AssertExpectations(t)
		})
	})

}

func TestPackageHolding(t *testing.T) {
	Convey("Given holding a package", t, func() {

		Convey("The package should be held", func() {
			r := new(MockRunner)
			r.On("Run", "apt-mark", []string{"hold", "apt"}).Return("", nil)
			runner = r
			err := holdPackage("apt")
			So(err, ShouldBeNil)
			r.Mock.AssertExpectations(t)
		})
	})

	Convey("Given unholding a package", t, func() {

		Convey("The package should be unheld", func() {
			r := new(MockRunner)
			r.On("Run", "apt-mark", []string{"unhold", "apt"}).Return("", nil)
			runner = r
			err := unholdPackage("apt")
			So(err, ShouldBeNil)
			r.Mock.AssertExpectations(t)
		})
	})

}

func TestSourceList(t *testing.T) {
	Convey("Given /etc/apt/sources.list exists", t, func() {

		Convey("There should be a source list", func() {
			packageList := []string{"deb http://us.archive.ubuntu.com/ubuntu/ trusty main restricted",
				"deb-src http://us.archive.ubuntu.com/ubuntu/ trusty main restricted"}
			r := new(MockRunner)
			r.On("Run", "grep", []string{"-h", "^deb", "/etc/apt/sources.list", "/etc/apt/sources.list.d/*"}).Return(strings.Join(packageList, "\n"), nil)
			runner = r
			sourcesList := getSourcesList()
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
	Convey("Given I want to view all installed packages", t, func() {
		Convey("It returns a list of all installed packages", func() {
			r := new(MockRunner)
			r.On("Run", "dpkg", []string{"--get-selections"}).Return("apt\u0009installed", nil)
			runner = r
			packages := buildInstalledPackageList()
			So(packages[0], ShouldEqual, "apt")
			So(len(packages), ShouldEqual, 1)
			r.Mock.AssertExpectations(t)
		})
	})

}

func TestUpdatingThePackageList(t *testing.T) {
	Convey("Given I want to have the latest source list", t, func() {

		Convey("apt-update gets run", func() {
			r := new(MockRunner)
			r.On("Run", "apt-get", []string{"update"}).Return("", nil)
			runner = r
			err := updatePackageLists()
			So(err, ShouldBeNil)
			r.Mock.AssertExpectations(t)
		})
	})

}

func TestUpdateCounts(t *testing.T) {
	Convey("Given there are security and regular updates", t, func() {

		Convey("The number of security and regular updates is > 0", func() {
			r := new(MockRunner)
			r.On("Run", "/usr/lib/update-notifier/apt-check", []string{}).Return("1;2", nil)
			runner = r
			updates := updateCounts()
			So(updates.Regular, ShouldEqual, 1)
			So(updates.Security, ShouldEqual, 2)
			r.Mock.AssertExpectations(t)
		})

		Convey("There are no security updates", func() {
			r := new(MockRunner)
			r.On("Run", "/usr/lib/update-notifier/apt-check", []string{}).Return("2;0", nil)
			runner = r
			updates := updateCounts()
			So(updates.Regular, ShouldEqual, 2)
			So(updates.Security, ShouldEqual, 0)
			r.Mock.AssertExpectations(t)
		})

	})
}
