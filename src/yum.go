package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sysward_agent/src/logging"
)

type CentosPackageManager struct{}

func (pm CentosPackageManager) UpdatePackage(pkg string) error {
	out, err := runner.Run("apt-get",
		"-y",
		"install", pkg)
	if os.Getenv("DEBUG") == "true" {
		debugMsg := strings.Join([]string{"apt-get",
			"-y",
			"install", pkg}, " ")
		logging.LogMsg("Command: " + debugMsg)
	}
	logging.LogMsg(string(out))
	if err != nil {
		err = errors.New(string(out) + err.Error())
	}
	return err
}

func (pm CentosPackageManager) HoldPackage(pkg string) error {
	out, err := runner.Run("apt-mark", "hold", pkg)
	logging.LogMsg(string(out))
	if err != nil {
		err = errors.New(string(out) + err.Error())
	}
	return err
}

func (pm CentosPackageManager) UnholdPackage(pkg string) error {
	out, err := runner.Run("apt-mark", "unhold", pkg)
	logging.LogMsg(string(out))
	if err != nil {
		err = errors.New(string(out) + err.Error())
	}
	return err
}

func (pm CentosPackageManager) GetSourcesList() []Source {
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

func (pm CentosPackageManager) GetChangelog(package_name string) string {
	changelog, _ := runner.Run("apt-get", "changelog", package_name)
	changelog_encoded := base64.StdEncoding.EncodeToString([]byte(changelog))
	return changelog_encoded
}

func (pm CentosPackageManager) BuildInstalledPackageList() []string {
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

func (pm CentosPackageManager) BuildPackageList() []OsPackage {
	out, err := runner.Run("python", "yum.py")
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

func (pm CentosPackageManager) UpdatePackageLists() error {
	_, err := runner.Run("apt-get", "update")
	return err
}

func (pm CentosPackageManager) UpdateCounts() Updates {
	out, err := runner.Run("yum", "list", "updates", "-q")
	if err != nil {
		panic(err)
	}
	regularOutput := strings.Split(string(out), "\n")[1:] // skip first line which is 'Updated Packages'
	out, err = runner.Run("yum", "list", "updates", "-q", "--security")
	securityOutput := strings.Split(string(out), "\n")
	regular := len(regularOutput)
	security := len(securityOutput)
	return Updates{regular, security}
}
