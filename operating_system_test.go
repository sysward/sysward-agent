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

//[root@centos7-sysward ~]# needs-restarting -r
//No core libraries or services have been updated.
//Reboot is probably not necessary.
//[root@centos7-sysward ~]# echo $?
//0
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
		f.On("FileExists", "/var/run/reboot-required").
			Return(false)
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
		f.On("FileExists", "/usr/lib/update-notifier/apt-check").Return(true)
		f.On("FileExists", "/usr/bin/python").Return(true)
		f.On("FileExists", "/usr/lib/python2.7/dist-packages/apt/__init__.py").Return(true)
		f.On("FileExists", "/usr/lib/python3/dist-packages/apt/__init__.py").Return(false)
		fileReader = f
		So(func() { checkPreReqs() }, ShouldNotPanic)
		// here
		f.Mock.AssertExpectations(t)
	})

	Convey("Given pre-req's aren't installed", t, func() {
		agent := NewAgent()
		agent.linux = "debian"
		r := new(MockRunner)
		f := new(MockReader)
		f.On("FileExists", "/usr/lib/update-notifier/apt-check").Return(false)
		f.On("FileExists", "/usr/bin/python").Return(false)
		f.On("FileExists", "/usr/lib/python2.7/dist-packages/apt/__init__.py").Return(false)
		f.On("FileExists", "/usr/lib/python3/dist-packages/apt/__init__.py").Return(false)
		r.On("Run", "apt-get", []string{"update"}).Return("", nil)
		r.On("Run", "apt-get", []string{"install", "update-notifier", "-y"}).Return("", nil)
		r.On("Run", "apt-get", []string{"install", "python", "-y"}).Return("", nil)
		r.On("Run", "apt-get", []string{"install", "python-apt", "-y"}).Return("", nil)
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
	r.On("Run", "lsb_release", []string{"-d"}).Return("Description:    Ubuntu 14.04 LTS", nil)
	r.On("Run", "grep", []string{"MemTotal", "/proc/meminfo"}).Return("MemTotal:        1017764 kB", nil)
	r.On("Run", "grep", []string{"name", "/proc/cpuinfo"}).Return("model name      : Intel(R) Core(TM) i7-4850HQ CPU @ 2.30GHz", nil)
	runner = r

	f := new(MockReader)
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
