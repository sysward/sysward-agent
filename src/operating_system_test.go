package main

import (
	"errors"
	"fmt"
	"os"

	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type TestReader struct{}
type TestErrorReader struct{}

func (r TestReader) ReadFile(path string) ([]byte, error) {
	if path == "/sys/class/dmi/id/product_uuid" {
		return []byte("UUID"), nil
	}
	return nil, nil
}

func (r TestErrorReader) ReadFile(path string) ([]byte, error) {
	if path == "/sys/class/dmi/id/product_uuid" {
		return nil, errors.New("fail")
	}
	return nil, nil
}

func TestSystemUid(t *testing.T) {
	Convey("Given /sys/class/dmi/id/product_uuid exists", t, func() {
		file_reader = TestReader{}
		Convey("Then I should get a UID", func() {
			uid := getSystemUID()
			So(uid, ShouldEqual, "UUID")
		})

	})
	Convey("Given /sys/class/dmi/id/product_uuid doesnt exist", t, func() {
		file_reader = TestErrorReader{}
		Convey("Then I should panic", func() {
			So(func() { getSystemUID() }, ShouldPanic)
		})

	})

}

func TestPrereqs(t *testing.T) {

	Convey("Given pre-req's are installed", t, nil)

	Convey("Given pre-req's aren't installed", t, nil)

}

func TestPrivilegeEscalation(t *testing.T) {

	Convey("Given I have sudo acccess", t, nil)

	Convey("Given I don't have sudo access", t, nil)

	Convey("Given I need to be root", t, func() {

		Convey("I am root", func() {

		})

		Convey("I am not root", func() {

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

func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	cmd := os.Args[3]
	defer os.Exit(0)

	if cmd == "python" && os.Args[4] == "trex.py" {
		fmt.Println(`[{"name": "apt", "section": "admin", "priority": "important", "current_version": "1.0.1ubuntu2", "security": true, "candidate_version": "1.0.1ubuntu2.1"}]`)
	} else if cmd == "apt-get" && os.Args[4] == "install" {
		if os.Args[6] == "apt-held" {
			os.Exit(-1)
		}
	} else if cmd == "apt-mark" {
		// nothing

	} else if cmd == "apt-get" && os.Args[4] == "update" {
		// nothing
	} else if cmd == "grep" && os.Args[4] == "-h" {
		src := `deb http://us.archive.ubuntu.com/ubuntu/ trusty main restricted
deb-src http://us.archive.ubuntu.com/ubuntu/ trusty main restricted`
		fmt.Println(src)
	} else if cmd == "dpkg" && os.Args[4] == "--get-selections" {
		fmt.Println("apt\u0009install")
	} else if cmd == "/usr/lib/update-notifier/apt-check" {
		fmt.Println("1;2")
	} else {
		fmt.Println(os.Args)
		for index, arg := range os.Args {
			fmt.Println(fmt.Sprintf("arg[%d]: %s", index, string(arg)))
		}
		os.Exit(-1)
	}
}
