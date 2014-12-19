package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sysward_agent/src/logging"
)

type DebianPackageManager struct{}

func (pm DebianPackageManager) UpdatePackage(pkg string) error {
	out, err := runner.Run("apt-get",
		"install",
		"-y",
		"-o",
		fmt.Sprintf("Dpkg::Options::=--force-confdef"),
		"-o",
		fmt.Sprintf("Dpkg::Options::=--force-confold"),
		pkg,
	)
	if os.Getenv("DEBUG") == "true" {
		logging.LogMsg(string(out))
	}
	if err != nil {
		err = errors.New(string(out) + err.Error())
	}
	return err
}

func (pm DebianPackageManager) HoldPackage(pkg string) error {
	out, err := runner.Run("apt-mark", "hold", pkg)
	logging.LogMsg(string(out))
	if err != nil {
		err = errors.New(string(out) + err.Error())
	}
	return err
}

func (pm DebianPackageManager) UnholdPackage(pkg string) error {
	out, err := runner.Run("apt-mark", "unhold", pkg)
	logging.LogMsg(string(out))
	if err != nil {
		err = errors.New(string(out) + err.Error())
	}
	return err
}

func (pm DebianPackageManager) GetSourcesList() []Source {
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

func (pm DebianPackageManager) GetChangelog(package_name string) string {
	changelog, _ := runner.Run("apt-get", "changelog", package_name)
	changelog_encoded := base64.StdEncoding.EncodeToString([]byte(changelog))
	return changelog_encoded
}

func (pm DebianPackageManager) BuildInstalledPackageList() []string {
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

func (pm DebianPackageManager) BuildPackageList() []OsPackage {
	out, err := runner.Run("python", "trex.py")
	if err != nil {
		panic(err)
	}
	var packages []OsPackage

	err = json.Unmarshal([]byte(out), &packages)
	if err != nil {
		panic(err)
	}

	return packages
}

func (pm DebianPackageManager) UpdatePackageLists() error {
	_, err := runner.Run("apt-get", "update")
	return err
}

func (pm DebianPackageManager) UpdateCounts() Updates {
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
