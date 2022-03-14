package main

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"

	"github.com/sysward/sysward-agent/logging"
)

func rebootRequired() bool {
	if agent.linux == "debian" {
		return fileReader.FileExists("/var/run/reboot-required")
	}
	if agent.linux == "centos" {
		out, _ := runner.Run("needs-restarting", "-r")
		return strings.Contains(out,"Reboot is required")
	}
	if agent.linux == "suse" {
		out, _ := runner.Run("zypper", "ps", "-s")
		return strings.Contains(out, "You may wish to restart these processes.")
	}
	return false
}

func getSystemUID() string {
	if fileReader.FileExists("/opt/sysward/bin/uid") {
		readUID, err := fileReader.ReadFile("/opt/sysward/bin/uid")
		if err != nil {
			logging.LogMsg(fmt.Sprintf("Error reading UID file: %s", err))
		} else {
			return string(readUID)
		}
	}
	interface_list, _ := net.Interfaces()
	var uuid []string

	for _, ifdev := range interface_list {
		addToList := true
		for _, s := range uuid {
			if strings.Contains(ifdev.Name, "dummy") {
				addToList = false
				continue
			}
			if os.Getenv("DEBUG") == "true" {
				logging.LogMsg("Name: " + ifdev.Name + " <-> HWADDR: " + ifdev.HardwareAddr.String())
			}
			if s == ifdev.HardwareAddr.String() {
				addToList = false
			}
		}
		if addToList {
			uuid = append(uuid, ifdev.HardwareAddr.String())
		}
	}

	uid := strings.Join(uuid, ".")
	if os.Getenv("DEBUG") == "true" {
		logging.LogMsg("UID: " + uid)
	}
	trimmedUID := strings.TrimSpace(string(uid))
	fw := SyswardFileWriter{}
	os.Create("/opt/sysward/bin/uid")
	fw.AppendToFile("/opt/sysward/bin/uid", trimmedUID)
	return trimmedUID
}

func checkPreReqs() {
	if agent.linux == "debian" {
		if !fileReader.FileExists("/usr/lib/update-notifier/apt-check") {
			fmt.Println("update notifier not found, installing")
			_, err := runner.Run("apt-get", "update")
			out, err := runner.Run("apt-get", "install", "update-notifier", "-y")
			if err != nil {
				panic(err)
			}
			logging.LogMsg(out)
		}

		if !fileReader.FileExists("/usr/bin/python") {
			fmt.Println("python not found, installing")
			_, err := runner.Run("apt-get", "update")
			out, err := runner.Run("apt-get", "install", "python", "-y")
			if err != nil {
				panic(err)
			}
			logging.LogMsg(out)
		}

		if !fileReader.FileExists("/usr/lib/python2.7/dist-packages/apt/__init__.py") &&
			!fileReader.FileExists("/usr/lib/python3/dist-packages/apt/__init__.py") {
			fmt.Println("python-apt not found, installing")
			_, err := runner.Run("apt-get", "update")
			out, err := runner.Run("apt-get", "install", "python-apt", "-y")
			if err != nil {
				panic(err)
			}
			logging.LogMsg(out)
		}
	} else if agent.linux == "suse" {
		if !fileReader.FileExists("/usr/bin/lsb_release") {
			fmt.Println("lsb_release not found, installing")
			out, err := runner.Run("zypper", "install", "-y", "lsb-release")
			if err != nil {
				panic(err)
			}
			logging.LogMsg(out)
		}
	} else if agent.linux == "centos" {
		if !fileReader.FileExists("/usr/bin/lsb_release") {
			fmt.Println("lsb_release not found, installing")
			out, err := runner.Run("yum", "install", "-y", "redhat-lsb-core")
			if err != nil {
				panic(err)
			}
			logging.LogMsg(out)
		}

		if !fileReader.FileExists("/usr/bin/needs-restarting") {
			fmt.Println("needs-restarting not found, installing")
			out, err := runner.Run("yum", "install", "-y", "yum-utils")
			if err != nil {
				panic(err)
			}
			logging.LogMsg(out)
		}

		if !fileReader.FileExists("/usr/bin/wget") {
			fmt.Println("wget not found, installing")
			out, err := runner.Run("yum", "install", "-y", "wget")
			if err != nil {
				panic(err)
			}
			logging.LogMsg(out)
		}

		if !fileReader.FileExists("/etc/yum/pluginconf.d/versionlock.conf") {
			fmt.Println("yum-plugin-versionlock.noarch not found, installing")
			out, err := runner.Run("yum", "install", "-y", "yum-plugin-versionlock.noarch")
			if err != nil {
				panic(err)
			}
			logging.LogMsg(out)
		}

	}
}

func verifyRoot() string {
	// cant use user.Current() because we're cross compiling and no cgo
	usr, err := runner.Run("whoami")
	if err != nil {
		panic(fmt.Sprintf("Return: %s | Error: %s", usr, err))
	}

	user := strings.TrimSpace(string(usr))
	if user != "root" {
		panic("SysWard agent must be run as root or as sudo.")
	}
	logging.LogMsg("root verified")
	return user
}

func getOsInformation() OperatingSystem {
	out, err := runner.Run("lsb_release", "-d")
	if err != nil {
		panic(err)
	}
	output := strings.Split(strings.TrimSpace(out), ":")
	tmp := strings.Split(strings.TrimSpace(output[1]), " ")
	hostname, err := os.Hostname()

	osName := tmp[0]
	osVersion := tmp[1]

	if len(tmp) > 2 {
		switch tmp[0] {
		case "CentOS":
			osVersion = tmp[2]
		case "Debian":
			osVersion = tmp[2]
		default:
		}

	}

	cpu_information := CPUInformation{getCPUName(), runtime.NumCPU()}
	memory_information := MemoryInformation{getTotalMemory()}
	return OperatingSystem{
		Name:              osName,
		UID:               getSystemUID(),
		Version:           osVersion,
		Interfaces:        getInterfaceInformation(),
		Hostname:          hostname,
		CPUInformation:    cpu_information,
		MemoryInformation: memory_information,
	}
}

func getTotalMemory() string {
	out, _ := runner.Run("grep", "MemTotal", "/proc/meminfo")
	t := strings.Split(out, ":")
	x := strings.TrimSpace(t[1])
	return x
}

func getCPUName() string {
	out, _ := runner.Run("grep", "name", "/proc/cpuinfo")
	t := strings.Split(out, ":")
	return strings.TrimSpace(t[1])
}

func getInterfaceInformation() []Interface {
	interface_list, _ := net.Interfaces()

	iflist := make([]Interface, len(interface_list))
	for index, ifdev := range interface_list {
		ips, _ := ifdev.Addrs()
		// skip interfaces that are from docker
		if strings.Contains(ifdev.Name, "veth") {
			logging.LogMsg(fmt.Sprintf("Skipping: %+v", ifdev))
			continue
		}
		ip_list := make([]string, len(ips))
		for c, ip := range ips {
			ip_list[c] = ip.String()
		}
		iflist[index] = Interface{ifdev.Name, ip_list, ifdev.HardwareAddr.String()}
	}
	return iflist
}
