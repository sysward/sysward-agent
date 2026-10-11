package main

import (
	"encoding/json"
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

//linux-vs8b:~ # zypper ps -s
//No processes using deleted files found.
//linux-vs8b:~ # echo $?
//0

//linux-vs8b:/opt/sysward/bin # zypper ps -s
//The following running processes use deleted files:
//
//PID  | PPID | UID | User       | Command                    | Service
//-----+------+-----+------------+----------------------------+-----------------
//361  | 1    | 0   | root       | systemd-journald (deleted) | systemd-journald
//529  | 1    | 0   | root       | auditd                     | auditd
//545  | 1    | 499 | messagebus | dbus-daemon                | dbus
//555  | 1    | 0   | root       | login                      |
//568  | 1    | 0   | root       | systemd-logind             | systemd-logind
//1303 | 1    | 0   | root       | cron                       | cron
//1308 | 1    | 0   | root       | systemd                    |
//1309 | 1308 | 0   | root       | (sd-pam)                   |
//1311 | 555  | 0   | root       | bash                       |
//1860 | 1    | 0   | root       | sshd                       | sshd
//2112 | 1860 | 0   | root       | sshd                       |
//2114 | 2112 | 0   | root       | bash                       |
//2623 | 1    | 0   | root       | zypper                     |
//
//You may wish to restart these processes.
//See 'man zypper' for information about the meaning of values in the above table.
//linux-vs8b:/opt/sysward/bin # echo $?
//0

//[root@centos7-sysward ~]# needs-restarting -r
//Core libraries or services have been updated:
//kernel -> 3.10.0-693.21.1.el7
//linux-firmware -> 20170606-58.gitc990aae.el7_4

//Reboot is required to ensure that your system benefits from these updates.

//More information:
//https://access.redhat.com/solutions/27943
//[root@centos7-sysward ~]# echo $?
//1

// [root@centos7-sysward ~]# needs-restarting -r
// No core libraries or services have been updated.
// Reboot is probably not necessary.
// [root@centos7-sysward ~]# echo $?
// 0
func TestRebootRequired(t *testing.T) {
	Convey("It has a pending reboot file", t, func() {
		agent := NewAgent()
		Convey("for debian flavors", func() {
			f := new(MockReader)
			fileReader = f
			agent.linux = "debian"
			f.On("FileExists", "/var/run/reboot-required").Return(true)
			So(rebootRequired(), ShouldEqual, true)
			f.Mock.AssertExpectations(t)
		})

		Convey("for centos flavors", func() {
			agent.linux = "centos"
			r := new(MockRunner)
			r.On("Run", "needs-restarting", []string{"-r"}).
				Return("Reboot is required.", nil)
			runner = r
			So(rebootRequired(), ShouldEqual, true)
			r.Mock.AssertExpectations(t)
		})

		Convey("for suse flavors", func() {
			agent.linux = "suse"
			r := new(MockRunner)
			r.On("Run", "zypper", []string{"ps", "-s"}).
				Return("You may wish to restart these processes.", nil)
			runner = r
			So(rebootRequired(), ShouldEqual, true)
			r.Mock.AssertExpectations(t)
		})
	})

	Convey("It doesnt have a pending reboot file", t, func() {
		agent := NewAgent()
		agent.linux = "debian"
		f := new(MockReader)
		f.On("FileExists", "/var/run/reboot-required").Return(false)
		f.On("FileExists", "/var/run/reboot-required.pkgs").Return(false)
		fileReader = f
		So(rebootRequired(), ShouldEqual, false)
		f.Mock.AssertExpectations(t)
	})
}

func TestPrereqs(t *testing.T) {

	Convey("Given pre-req's are installed", t, func() {
		f := new(MockReader)
		agent := NewAgent()
		agent.linux = "debian"
		fileReader = f
		r := new(MockRunner)
		fileReader = f
		runner = r
		So(func() { checkPreReqs() }, ShouldNotPanic)
		// here
		f.Mock.AssertExpectations(t)
	})

	Convey("Given pre-req's aren't installed", t, func() {
		agent := NewAgent()
		agent.linux = "debian"
		r := new(MockRunner)
		f := new(MockReader)
		fileReader = f
		runner = r
		// here
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
	r.On("Run", "grep", []string{"MemTotal", "/proc/meminfo"}).Return("MemTotal:        1017764 kB", nil)
	r.On("Run", "grep", []string{"name", "/proc/cpuinfo"}).Return("model name      : Intel(R) Core(TM) i7-4850HQ CPU @ 2.30GHz", nil)
	runner = r

	f := new(MockReader)
	f.On("ReadFile", "/etc/os-release").Return([]byte(`
NAME="Ubuntu"
VERSION="20.04.6 LTS (Focal Fossa)"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 20.04.6 LTS"
VERSION_ID="20.04"
HOME_URL="https://www.ubuntu.com/"
SUPPORT_URL="https://help.ubuntu.com/"
BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"
PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"
VERSION_CODENAME=focal
UBUNTU_CODENAME=focal
`), nil)
	f.On("ReadFile", "/opt/sysward/bin/uid").Return([]byte("uid123"), nil)
	f.On("FileExists", "/opt/sysward/bin/uid").Return(true)
	f.On("FileExists", "/usr/bin/python").Return(true)
	f.On("FileExists", "/usr/lib/python2.7/dist-packages/apt/__init__.py").Return(true)
	fileReader = f

	os := getOsInformation()
	Convey("Given I run lsb_release -a", t, func() {

		Convey("It should have an OS name", func() {
			So(os.Name, ShouldEqual, "Ubuntu")
		})

		Convey("It should have a UID", func() {
			So(os.UID, ShouldNotBeNil)
		})

		Convey("It should have an OS version", func() {
			So(os.Version, ShouldEqual, "20.04")
		})

		Convey("It should have an os-release ID and ID_LIKE", func() {
			So(os.ID, ShouldEqual, "ubuntu")
			So(os.IDLike, ShouldEqual, "debian")
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

func TestParseOsRelease(t *testing.T) {
	Convey("Given an os-release file", t, func() {

		Convey("RHEL style quoting, comments, blank lines and '=' in values", func() {
			info := parseOsRelease(`# this is a comment
NAME="Red Hat Enterprise Linux"
VERSION="9.4 (Plow)"

ID="rhel"
ID_LIKE="fedora"
VERSION_ID="9.4"
  # indented comment
not a key value line
=novalue
PRETTY_NAME='Red Hat Enterprise Linux 9.4 (Plow)'
HOME_URL="https://www.redhat.com/?a=b&c=d"
EMPTY=
`)
			So(info["NAME"], ShouldEqual, "Red Hat Enterprise Linux")
			So(info["ID"], ShouldEqual, "rhel")
			So(info["ID_LIKE"], ShouldEqual, "fedora")
			So(info["VERSION_ID"], ShouldEqual, "9.4")
			So(info["PRETTY_NAME"], ShouldEqual, "Red Hat Enterprise Linux 9.4 (Plow)")
			So(info["HOME_URL"], ShouldEqual, "https://www.redhat.com/?a=b&c=d")
			So(info["EMPTY"], ShouldEqual, "")
			So(len(info), ShouldEqual, 8)
		})

		Convey("Unquoted multi-word ID_LIKE stays intact", func() {
			info := parseOsRelease("ID=rocky\nID_LIKE=\"rhel centos fedora\"\n")
			So(info["ID"], ShouldEqual, "rocky")
			So(info["ID_LIKE"], ShouldEqual, "rhel centos fedora")
		})

		Convey("A lone quote is not stripped into nothing", func() {
			info := parseOsRelease("X=\"\n")
			So(info["X"], ShouldEqual, "\"")
		})
	})

	Convey("getOsInformation lowercases ID and keeps NAME/VERSION_ID", t, func() {
		r := new(MockRunner)
		r.On("Run", "grep", []string{"MemTotal", "/proc/meminfo"}).Return("MemTotal: 1 kB", nil)
		r.On("Run", "grep", []string{"name", "/proc/cpuinfo"}).Return("model name : cpu", nil)
		runner = r
		f := new(MockReader)
		f.On("ReadFile", "/etc/os-release").Return([]byte("NAME=\"openSUSE Leap\"\nVERSION_ID=\"15.5\"\nID=\"OpenSUSE-Leap\"\nID_LIKE=\"suse opensuse\"\n"), nil)
		f.On("ReadFile", "/opt/sysward/bin/uid").Return([]byte("uid123"), nil)
		f.On("FileExists", "/opt/sysward/bin/uid").Return(true)
		fileReader = f

		info := getOsInformation()
		So(info.Name, ShouldEqual, "openSUSE Leap")
		So(info.Version, ShouldEqual, "15.5")
		So(info.ID, ShouldEqual, "opensuse-leap")
		So(info.IDLike, ShouldEqual, "suse opensuse")
	})
}

func TestOperatingSystemJSONFields(t *testing.T) {
	Convey("The OS payload carries id and id_like", t, func() {
		b, err := json.Marshal(OperatingSystem{ID: "rhel", IDLike: "fedora"})
		So(err, ShouldBeNil)
		So(string(b), ShouldContainSubstring, `"id":"rhel"`)
		So(string(b), ShouldContainSubstring, `"id_like":"fedora"`)
	})

	Convey("The package payload carries source, arch and installed_evr", t, func() {
		b, err := json.Marshal(OsPackage{Source: "openssl", Arch: "amd64", InstalledEVR: "3.0.2-0ubuntu1.15"})
		So(err, ShouldBeNil)
		So(string(b), ShouldContainSubstring, `"source":"openssl"`)
		So(string(b), ShouldContainSubstring, `"arch":"amd64"`)
		So(string(b), ShouldContainSubstring, `"installed_evr":"3.0.2-0ubuntu1.15"`)
	})
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
