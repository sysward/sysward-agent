package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSystemUid(t *testing.T) {
	Convey("Given i have valid network interfaces with MACs", t, func() {
		Convey("Then I should get a UID", func() {
			So(getSystemUID(), ShouldNotBeNil)
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
		r.On("Run", "apt-get", []string{"update"}).Return("", nil)
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
			r.Mock.AssertExpectations(t)
		})

		Convey("I am not root", func() {
			r := new(MockRunner)
			r.On("Run", "whoami", []string{}).Return("notroot", nil)
			runner = r
			So(func() { verifyRoot() }, ShouldPanic)
			r.Mock.AssertExpectations(t)
		})

	})

	Convey("Give I am not root and don't have sudo access", t, nil)

}

func TestOSInformation(t *testing.T) {

	r := new(MockRunner)
	r.On("Run", "lsb_release", []string{"-d"}).Return("Description:    Ubuntu 14.04 LTS", nil)
	r.On("Run", "grep", []string{"MemTotal", "/proc/meminfo"}).Return("MemTotal:        1017764 kB", nil)
	r.On("Run", "grep", []string{"name", "/proc/cpuinfo"}).Return("model name      : Intel(R) Core(TM) i7-4850HQ CPU @ 2.30GHz", nil)
	runner = r

	os := getOsInformation()
	Convey("Given I run lsb_release -a", t, func() {

		Convey("It should have an OS name", func() {
			So(os.Name, ShouldEqual, "Ubuntu")
		})

		Convey("It should have a UID", func() {
			So(os.UID, ShouldNotBeNil)
		})

		Convey("It should have an OS version", func() {
			So(os.Version, ShouldEqual, "14.04")
		})

		Convey("It should have network interfaces", func() {
		})

		Convey("It should have a hostname", func() {
			So(os.Hostname, ShouldNotBeNil)
		})

		Convey("It should have CPU information", func() {
			So(os.CPUInformation.Name, ShouldEqual, "Intel(R) Core(TM) i7-4850HQ CPU @ 2.30GHz")
		})

		Convey("It should have Memory information", func() {
			So(os.MemoryInformation.Total, ShouldEqual, "1017764 kB")
		})

	})

	r.Mock.AssertExpectations(t)
}

func TestMemory(t *testing.T) {

	Convey("It should give me total memory", t, func() {
		r := new(MockRunner)
		r.On("Run", "grep", []string{"MemTotal", "/proc/meminfo"}).Return("MemTotal:        1017764 kB", nil)
		runner = r
		So(getTotalMemory(), ShouldEqual, "1017764 kB")
		r.Mock.AssertExpectations(t)
	})

}

func TestCPUInformation(t *testing.T) {

	Convey("It should give me the CPU name", t, func() {
		r := new(MockRunner)
		r.On("Run", "grep", []string{"name", "/proc/cpuinfo"}).Return("model name      : Intel(R) Core(TM) i7-4850HQ CPU @ 2.30GHz", nil)
		runner = r
		So(getCPUName(), ShouldEqual, "Intel(R) Core(TM) i7-4850HQ CPU @ 2.30GHz")
	})

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
