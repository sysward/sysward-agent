package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPackageUpdates(t *testing.T) {

	Convey("Given a package name", t, func() {

		Convey("The package should be upgraded", nil)

		Convey("The package should not upgrade if held", nil)

	})

}

func TestPackageHolding(t *testing.T) {

	Convey("Given holding a package", t, func() {

		Convey("The package should be held", nil)

	})

	Convey("Given unholding a package", t, func() {

		Convey("The package should be unheld", nil)

	})

}

func TestSourceList(t *testing.T) {

	Convey("Given /etc/apt/sources.list exists", t, func() {

		Convey("There should be a source list", nil)

	})

}

func TestInstalledPackages(t *testing.T) {

	Convey("Given I want to view all installed packages", t, func() {

		Convey("It returns a list of all installed packages", nil)

	})

}

func TestPackagesThatNeedUpdates(t *testing.T) {

	Convey("Given pending updates", t, func() {

		Convey("There should be a list of packages available for update", nil)

		Convey("There should be a list of security packages for update", nil)

	})

}

func TestUpdatingThePackageList(t *testing.T) {

	Convey("Given I want to have the latest source list", t, func() {

		Convey("apt-update gets run", nil)

	})

}

func TestUpdateCounts(t *testing.T) {

	Convey("Given there are security and regular updates", t, func() {

		Convey("The number of security and regular updates is > 0", nil)

	})

	Convey("Given there are only security updates", t, func() {

		Convey("The number of security updates should be > 0 and regular should == 0", nil)

	})

}
