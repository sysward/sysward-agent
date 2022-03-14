package main

import (
	"bitbucket.org/sysward/sysward-agent/logging"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
)

type CentosPackageManager struct{}

func (pm CentosPackageManager) UpdatePackage(pkg string) error {
	out, err := runner.Run("yum",
		"update",
		"-y", pkg)
	if os.Getenv("DEBUG") == "true" {
		debugMsg := strings.Join([]string{"yum",
			"update",
			"-y", pkg}, " ")
		logging.LogMsg("Command: " + debugMsg)
	}
	logging.LogMsg(string(out))
	if err != nil {
		err = errors.New(string(out) + err.Error())
	}
	return err
}

func (pm CentosPackageManager) HoldPackage(pkg string) error {
	out, err := runner.Run("yum", "versionlock", pkg)
	logging.LogMsg(string(out))
	if err != nil {
		err = errors.New(string(out) + err.Error())
	}
	return err
}

func (pm CentosPackageManager) UnholdPackage(pkg string) error {
	out, err := runner.Run("yum", "versionlock", "delete", pkg)
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
	changelog, _ := runner.Run("yum", "changelog", package_name)
	changelog_encoded := base64.StdEncoding.EncodeToString([]byte(changelog))
	return changelog_encoded
}

func (pm CentosPackageManager) BuildInstalledPackageList() []string {
	installed, _ := runner.Run("rpm", "-qa", "--queryformat", "%{name}\t%{version}\n")
	installed_arr := strings.Split(string(installed), "\n")[1:]
	packages := []string{}
	for _, line := range installed_arr {
		x := strings.Fields(line)
		if len(x) == 0 {
			continue
		}
		if x[0] == "" {
			continue
		}
		pkg_name := x[0]
		packages = append(packages, strings.TrimSpace(pkg_name))
	}
	return packages
}

func (pm CentosPackageManager) BuildPackageList() []OsPackage {
	out, err := runner.Run("python", "list_updates.py")
	var packages []OsPackage

	if out == "" {
		return packages
	}

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
	packages := pm.BuildPackageList()
	security := 0
	regular := 0
	for _, p := range packages {
		if p.Security {
			security++
		} else {
			regular++
		}
	}
	return Updates{regular, security}
}
