package main

import (
	"bitbucket.org/sysward/sysward-agent/logging"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
)

func getSystemUID() string {
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
	return strings.TrimSpace(string(uid))
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
	} else if agent.linux == "centos" {
		if !fileReader.FileExists("/usr/bin/lsb_release") {
			fmt.Println("lsb_release not found, installing")
			out, err := runner.Run("yum", "install", "-y", "redhat-lsb-core")
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

	cpu_information := CPUInformation{getCPUName(), runtime.NumCPU()}
	memory_information := MemoryInformation{getTotalMemory()}
	return OperatingSystem{tmp[0], getSystemUID(), tmp[1], getInterfaceInformation(), hostname, cpu_information, memory_information}
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
