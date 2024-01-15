package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"github.com/sysward/sysward-agent/logging"
	"os"
	"os/exec"
	"strings"
)

type CentosPackageManager struct {
	ForceYum bool
}

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
	// build list of security updates first
	pkgManager := "dnf"
	if pm.ForceYum {
		pkgManager = "yum"
	}
	cmd := exec.Command(pkgManager, "list", "updates", "--security")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	// Parse the command output to extract update information

	security := map[string]struct{}{}
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		security[fields[0]] = struct{}{}
	}

	cmd = exec.Command(pkgManager, "list", "updates")
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	// Parse the command output to extract update information
	var updates []OsPackage
	lines = strings.Split(out.String(), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		// Use rpm command to get the installed version
		rpmCmd := exec.Command("rpm", "-q", fields[0])
		rpmOut, err := rpmCmd.Output()
		if err != nil {
			continue // Skip if we can't get the installed version
		}
		installedVersion := strings.TrimSpace(string(rpmOut))

		_, isSecurity := security[fields[0]]

		// Create an Update struct and append it to the list
		updates = append(updates, OsPackage{
			Name:              fields[0],
			Current_version:   installedVersion,
			Candidate_version: fields[1],
			// The Priority and Section information are not available directly from dnf list updates command
			// If they are available from another source, you will need to adjust this part
			Priority: "N/A",
			Section:  fields[2],
			Security: isSecurity,
		})
	}

	return updates
}

func (pm CentosPackageManager) UpdatePackageLists() error {
	// NOOP
	return nil
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
