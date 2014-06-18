package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSystemUid(t *testing.T) {

	Convey("Given /sys/class/dmi/id/product_uuid exists", t, func() {

		Convey("Then I should get a UID", nil)

	})

}

func TestPrereqs(t *testing.T) {

	Convey("Given pre-req's are installed", t, nil)

	Convey("Given pre-req's aren't installed", t, nil)

}

func TestPrivilegeEscalation(t *testing.T) {

	Convey("Given I have sudo acccess", t, nil)

	Convey("Given I don't have sudo access", t, nil)

	Convey("Given I am root", t, nil)

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
