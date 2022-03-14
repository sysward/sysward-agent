package main

import (
	"encoding/base64"
	//	"encoding/json"
	"github.com/sysward/sysward-agent/logging"
	"errors"
	"fmt"
	"os"
	//	"strconv"
	"strings"
)

type ZypperPackageManager struct{}

func (pm ZypperPackageManager) UpdatePackage(pkg string) error {
	out, err := runner.Run("zypper",
		"--non-interactive",
		"up",
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

func (pm ZypperPackageManager) HoldPackage(pkg string) error {
	out, err := runner.Run("zypper", "al", pkg)
	logging.LogMsg(string(out))
	if err != nil {
		err = errors.New(string(out) + err.Error())
	}
	return err
}

func (pm ZypperPackageManager) UnholdPackage(pkg string) error {
	out, err := runner.Run("zypper", "rl", pkg)
	logging.LogMsg(string(out))
	if err != nil {
		err = errors.New(string(out) + err.Error())
	}
	return err
}

func (pm ZypperPackageManager) GetSourcesList() []Source {
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

func (pm ZypperPackageManager) GetChangelog(package_name string) string {
	changelog, _ := runner.Run("apt-get", "changelog", package_name)
	changelog_encoded := base64.StdEncoding.EncodeToString([]byte(changelog))
	return changelog_encoded
}

func (pm ZypperPackageManager) BuildInstalledPackageList() []string {
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

type ZypperPatches struct {
	Category string
	Severity string
	Status   string
	Summary  string
}

func (pm ZypperPackageManager) IsSecurityUpdate(patch_list []ZypperPatches, package_name string) bool {
	for _, p := range patch_list {
		if !strings.Contains(p.Summary, package_name) {
			continue
		}
		return p.Category == "security"
	}
	return false
}

func (pm ZypperPackageManager) GetPatchList() []ZypperPatches {
	out, err := runner.Run("zypper", "list-patches")
	if err != nil {
		panic(err)
	}

	patch_list := []ZypperPatches{}

	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, "Update |") {
			continue
		}
		parts := strings.Split(line, "|")
		tmp := ZypperPatches{
			Category: strings.TrimSpace(parts[2]),
			Severity: strings.TrimSpace(parts[3]),
			Status:   strings.TrimSpace(parts[4]),
			Summary:  strings.TrimSpace(parts[5]),
		}

		patch_list = append(patch_list, tmp)
		if os.Getenv("DEBUG") == "true" {
			fmt.Printf("%+v\n", tmp)
		}
	}

	return patch_list
}

func (pm ZypperPackageManager) BuildPackageList() []OsPackage {
	out, err := runner.Run("zypper", "list-updates")
	if err != nil {
		panic(err)
	}

	patch_list := pm.GetPatchList()

	var packages []OsPackage

	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, "v |") {
			continue
		}
		parts := strings.Split(line, "|")
		pkg := OsPackage{
			Name:              strings.TrimSpace(parts[2]),
			Current_version:   strings.TrimSpace(parts[3]),
			Candidate_version: strings.TrimSpace(parts[4]),
			Security:          false,
		}
		pkg.Security = pm.IsSecurityUpdate(patch_list, pkg.Name)
		packages = append(packages, pkg)
		if os.Getenv("DEBUG") == "true" {
			fmt.Println(fmt.Sprintf("%+v", pkg))
		}
	}

	return packages
}

func (pm ZypperPackageManager) UpdatePackageLists() error {
	_, err := runner.Run("apt-get", "update")
	return err
}

func (pm ZypperPackageManager) UpdateCounts() Updates {
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
