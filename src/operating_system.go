package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func getSystemUID() string {
	uid, err := file_reader.ReadFile("/sys/class/dmi/id/product_uuid")
	fmt.Println(uid)
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(uid))
}

func checkPreReqs() {

	if _, err := os.Stat("/usr/lib/update-notifier/apt-check"); os.IsNotExist(err) {
		fmt.Println("update notifier not found, installing")
		out, err := exec.Command("apt-get", "install", "update-notifier", "-y").Output()
		if err != nil {
			panic(err)
		}
		fmt.Println(string(out))
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
		panic("patchasaurus client must be run as root.")
	}
	return user
}

func getOsInformation() OperatingSystem {
	out, err := exec.Command("lsb_release", "-d").Output()
	if err != nil {
		panic(err)
	}
	output := strings.Split(strings.TrimSpace(string(out)), ":")
	tmp := strings.Split(strings.TrimSpace(output[1]), " ")
	hostname, err := os.Hostname()

	cpu_information := CPUInformation{getCPUName(), runtime.NumCPU()}
	memory_information := MemoryInformation{getTotalMemory()}
	return OperatingSystem{tmp[0], getSystemUID(), tmp[1], getInterfaceInformation(), hostname, cpu_information, memory_information}
}

func getTotalMemory() string {
	out, _ := exec.Command("grep", "MemTotal", "/proc/meminfo").Output()
	t := strings.Split(string(out), ":")
	x := strings.TrimSpace(t[1])
	return x
}

func getCPUName() string {
	out, _ := exec.Command("grep", "name", "/proc/cpuinfo").Output()
	t := strings.Split(string(out), ":")
	return strings.TrimSpace(t[1])
}

func getInterfaceInformation() []Interface {
	interface_list, _ := net.Interfaces()

	iflist := make([]Interface, len(interface_list))
	for index, ifdev := range interface_list {
		ips, _ := ifdev.Addrs()
		ip_list := make([]string, len(ips))
		for c, ip := range ips {
			ip_list[c] = ip.String()
		}
		iflist[index] = Interface{ifdev.Name, ip_list, ifdev.HardwareAddr.String()}
	}
	return iflist
}
