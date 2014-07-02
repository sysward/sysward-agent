package main

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSystemUid(t *testing.T) {
	Convey("Given /sys/class/dmi/id/product_uuid exists", t, func() {
		Convey("Then I should get a UID", func() {
			r := new(MockReader)
			r.On("ReadFile", "/sys/class/dmi/id/product_uuid").Return([]byte("UUID"), nil)
			file_reader = r
			So(getSystemUID(), ShouldEqual, "UUID")
			r.Mock.AssertExpectations(t)
		})

	})
	Convey("Given /sys/class/dmi/id/product_uuid doesnt exist", t, func() {
		Convey("Then I should panic", func() {
			r := new(MockReader)
			r.On("ReadFile", "/sys/class/dmi/id/product_uuid").Return([]byte{}, errors.New("fail"))
			file_reader = r
			So(func() { getSystemUID() }, ShouldPanic)
			r.Mock.AssertExpectations(t)
		})

	})

}

func TestPrereqs(t *testing.T) {

	Convey("Given pre-req's are installed", t, func() {
		f := new(MockReader)
		f.On("FileExists", "/usr/lib/update-notifier/apt-check").Return(true)
		file_reader = f
		So(func() { checkPreReqs() }, ShouldNotPanic)
		f.Mock.AssertExpectations(t)
	})

	Convey("Given pre-req's aren't installed", t, func() {
		r := new(MockRunner)
		f := new(MockReader)
		f.On("FileExists", "/usr/lib/update-notifier/apt-check").Return(false)
		r.On("Run", "apt-get", []string{"install", "update-notifier", "-y"}).Return("", nil)
		file_reader = f
		runner = r
		checkPreReqs()
		f.Mock.AssertExpectations(t)
		r.Mock.AssertExpectations(t)
	})

}

func TestPrivilegeEscalation(t *testing.T) {
	Convey("Given I have sudo acccess", t, nil)

	Convey("Given I don't have sudo access", t, nil)

	Convey("Given I need to be root", t, func() {

		Convey("I am root", func() {
			r := new(MockRunner)
			r.On("Run", "whoami", []string{}).Return("root", nil)
			runner = r
			So(verifyRoot(), ShouldEqual, "root")
		})

		Convey("I am not root", func() {
			r := new(MockRunner)
			r.On("Run", "whoami", []string{}).Return("notroot", nil)
			runner = r
			So(func() { verifyRoot() }, ShouldPanic)
		})

	})

	Convey("Give I am not root and don't have sudo access", t, nil)

}

func TestOSInformation(t *testing.T) {

	Convey("Given I run lsb_release -a", t, func() {

		Convey("It should have an OS name", nil)

		Convey("It should have a UID", nil)

		Convey("It should have an OS version", nil)

		Convey("It should have network interfaces", nil)

		Convey("It should have a hostname", nil)

		Convey("It should have CPU information", nil)

		Convey("It should have Memory information", nil)

	})

}

func TestMemory(t *testing.T) {

	Convey("It should give me total memory", t, nil)

}

func TestCPUInformation(t *testing.T) {

	Convey("It should give me the CPU name", t, nil)

}

func TestInterfaceInformation(t *testing.T) {

	Convey("Given it has an interface on eth0", t, func() {

		Convey("It should give me an interface name", nil)

		Convey("It should have a MAC address", nil)

		Convey("Given it has one IP", func() {

			Convey("It should have a single IP", nil)

		})

		Convey("Given it has multiple IPs", func() {

			Convey("It should have multiple IPs", nil)

		})

	})

}
