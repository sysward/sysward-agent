package main

import (
	"encoding/base64"
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
)

func updatePackage(pkg string) error {
	out, err := runner.Run("apt-get", "install", "-y", pkg)
	logMsg(string(out))
	return err
}

func holdPackage(pkg string) error {
	out, err := runner.Run("apt-mark", "hold", pkg)
	logMsg(string(out))
	return err
}

func unholdPackage(pkg string) error {
	out, err := runner.Run("apt-mark", "unhold", pkg)
	logMsg(string(out))
	return err
}

func getSourcesList() []Source {
	out, _ := runner.Run("grep", "-h", "^deb", "/etc/apt/sources.list", "/etc/apt/sources.list.d/*")
	out_arr := strings.Split(strings.TrimSpace(string(out)), "\n")
	sources := make([]Source, len(out_arr))
	for index, o := range out_arr {
		x := strings.Split(o, " ")
		src := false
		if x[0] == "deb-src" {
			src = true
		}
		sources[index] = Source{x[1], src, x[2:]}
	}
	return sources
}

func getChangelog(package_name string) string {
	changelog, _ := exec.Command("apt-get", "changelog", package_name).Output()
	changelog_encoded := base64.StdEncoding.EncodeToString(changelog)
	return changelog_encoded
}

func buildInstalledPackageList() []string {
	installed, _ := runner.Run("dpkg", "--get-selections")
	installed_arr := strings.Split(string(installed), "\n")
	packages := []string{}
	for _, line := range installed_arr {
		x := strings.Split(line, "\u0009")
		if x[0] == "" {
			continue
		}
		packages = append(packages, x[0])
	}
	return packages
}

func buildPackageList() []OsPackage {
	out, err := runner.Run("python", "trex.py")
	if err != nil {
		panic(err)
	}
	var packages []OsPackage

	err = json.Unmarshal(out, &packages)
	if err != nil {
		panic(err)
	}

	return packages
}

func updatePackageLists() error {
	_, err := runner.Run("apt-get", "update")
	return err
}

func updateCounts() Updates {
	out, err := runner.Run("/usr/lib/update-notifier/apt-check")
	if err != nil {
		panic(err)
	}
	output := strings.TrimSpace(string(out))
	ups := strings.Split(output, ";")
	regular, err := strconv.Atoi(ups[0])
	if err != nil {
		panic(err)
	}
	security, err := strconv.Atoi(ups[1])
	if err != nil {
		panic(err)
	}
	return Updates{regular, security}
}
