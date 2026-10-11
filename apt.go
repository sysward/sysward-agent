package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/sysward/sysward-agent/logging"
	"os"
	//	"strconv"
	"strings"
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

// dpkgQueryFormat asks dpkg-query for every field BuildPackageList needs in
// one call: installed version, source package, architecture, priority, section.
const dpkgQueryFormat = "${Version}\t${source:Package}\t${Architecture}\t${Priority}\t${Section}\n"

type dpkgPackageInfo struct {
	Version  string
	Source   string
	Arch     string
	Priority string
	Section  string
}

// parseDpkgQuery parses dpkgQueryFormat output. Only the first line is used
// (a Multi-Arch: same package installed for several architectures yields one
// line per architecture). An empty source falls back to the package name.
func parseDpkgQuery(name string, out []byte) dpkgPackageInfo {
	line := strings.SplitN(strings.TrimLeft(string(out), "\n"), "\n", 2)[0]
	fields := strings.Split(line, "\t")
	get := func(i int) string {
		if i < len(fields) {
			return strings.TrimSpace(fields[i])
		}
		return ""
	}
	info := dpkgPackageInfo{
		Version:  get(0),
		Source:   get(1),
		Arch:     get(2),
		Priority: get(3),
		Section:  get(4),
	}
	if info.Source == "" {
		info.Source = strings.SplitN(name, ":", 2)[0]
	}
	return info
}

func (pm DebianPackageManager) BuildPackageList() []OsPackage {
	var packages []OsPackage

	out, err := runner.RunBytes("apt-get", "-s", "upgrade")
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	// Parse the output
	lines := bytes.Split(out, []byte("\n"))
	for _, line := range lines {
		parts := bytes.Fields(line)
		if len(parts) < 5 {
			continue
		}

		if !bytes.Contains(line, []byte("Inst")) {
			continue
		}

		// Parse the package information
		name := strings.Split(string(parts[1]), "/")[0]

		dpkgOut, err := runner.RunBytes("dpkg-query", "-W", "-f="+dpkgQueryFormat, name)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		availableVersion := strings.Replace(string(parts[3]), "(", "", -1)
		info := parseDpkgQuery(name, dpkgOut)

		isSecurity := bytes.Contains(parts[4], []byte("Security"))

		pkg := OsPackage{Name: name,
			Current_version:   info.Version,
			Candidate_version: availableVersion,
			Priority:          info.Priority,
			Section:           info.Section,
			Security:          isSecurity,
			Source:            info.Source,
			Arch:              info.Arch,
			InstalledEVR:      info.Version,
		}
		packages = append(packages, pkg)
	}

	return packages
}

func (pm DebianPackageManager) UpdatePackageLists() error {
	_, err := runner.Run("apt-get", "update")
	return err
}

func (pm DebianPackageManager) UpdateCounts() Updates {
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
