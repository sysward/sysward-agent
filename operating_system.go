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
		return strings.Contains(out, "Reboot is required")
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

	} else if agent.linux == "centos" {
		if !fileReader.FileExists("/usr/bin/dnf") {
			fmt.Println("dnf not found, installing")

			release, err := fileReader.ReadFile("/etc/os-release")
			if err != nil {
				logging.LogMsg("Error reading /etc/os-release: " + err.Error())
			}

			if !strings.Contains(string(release), "Amazon Linux") {
				out, err := runner.Run("yum", "install", "-y", "dnf")
				if err != nil {
					panic(err)
				}
				logging.LogMsg(out)
			} else {
				logging.LogMsg("Skipping dnf install, using amazon linux")
			}
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
				if strings.ContainsAny(out, "Unable to find a match: yum-plugin-versionlock.noarch") {
					var err2 error
					out, err2 = runner.Run("yum", "install", "-y", "python3-dnf-plugin-versionlock")
					if err2 != nil {
						panic(err)
					}
				} else {
					panic(err)
				}
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
	osRelease, err := fileReader.ReadFile("/etc/os-release")
	if err != nil {
		panic(err)
	}
	osInfo := map[string]string{}
	for _, line := range strings.Split(string(osRelease), "\n") {
		if len(line) == 0 {
			continue
		}
		kv := strings.Split(line, "=")
		osInfo[kv[0]] = strings.Trim(kv[1], "\"")
	}

	hostname, err := os.Hostname()

	cpu_information := CPUInformation{getCPUName(), runtime.NumCPU()}
	memory_information := MemoryInformation{getTotalMemory()}
	return OperatingSystem{
		Name:              osInfo["NAME"],
		UID:               getSystemUID(),
		Version:           osInfo["VERSION_ID"],
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
