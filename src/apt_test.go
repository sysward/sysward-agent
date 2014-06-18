package main

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"os/exec"
	"testing"
)

type TestRunner struct{}

func (r TestRunner) Run(command string, args ...string) ([]byte, error) {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	out, err := cmd.CombinedOutput()
	return out, err
}

func TestPackagesThatNeedUpdates(t *testing.T) {
	runner = TestRunner{}
	Convey("Given pending updates", t, func() {

		Convey("There should be a list of packages available for update", func() {
			out, _ := runner.Run("echo", "hello")
			fmt.Println(string(out))
		})

		Convey("There should be a list of security packages for update", nil)

	})

}

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

func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)
	fmt.Println("testing helper process")
}
