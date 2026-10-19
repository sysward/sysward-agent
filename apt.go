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

type aptInstLine struct {
	Name      string
	Candidate string
	Security  bool
}

// parseAptInstLine parses one "Inst" line of `apt-get -s upgrade`:
//
//	Inst libssl3 [3.0.2-0ubuntu1.15] (3.0.2-0ubuntu1.18 Ubuntu:22.04/jammy-updates, Ubuntu:22.04/jammy-security [amd64])
//
// The parenthesised part holds the candidate version, then one or more
// comma-separated "Label:Version/Archive" releases it is available from,
// then the architecture. The "[old]" part is absent for a new install.
func parseAptInstLine(line string) (aptInstLine, bool) {
	fields := strings.Fields(line)
	if len(fields) < 3 || fields[0] != "Inst" {
		return aptInstLine{}, false
	}
	open := strings.Index(line, "(")
	end := strings.LastIndex(line, ")")
	if open < 0 || end < open {
		return aptInstLine{}, false
	}
	inner := strings.TrimSpace(line[open+1 : end])
	if i := strings.LastIndex(inner, " ["); i >= 0 {
		inner = inner[:i]
	}
	parts := strings.SplitN(inner, " ", 2)
	inst := aptInstLine{
		Name:      strings.Split(fields[1], "/")[0],
		Candidate: parts[0],
	}
	if len(parts) == 2 {
		for _, release := range strings.Split(parts[1], ",") {
			if isAptSecurityRelease(strings.TrimSpace(release)) {
				inst.Security = true
			}
		}
	}
	return inst, true
}

// isAptSecurityRelease reports whether a "Label:Version/Archive" release
// is a security archive: Ubuntu "Ubuntu:22.04/jammy-security", Debian 11+
// "Debian-Security:12/stable-security" or "Debian:12/bookworm-security", and
// pre-11 Debian whose label alone says so ("Debian-Security:10/oldstable").
func isAptSecurityRelease(release string) bool {
	release = strings.ToLower(release)
	label, archive := release, release
	if i := strings.Index(release, ":"); i >= 0 {
		label = release[:i]
	}
	if i := strings.LastIndex(release, "/"); i >= 0 {
		archive = release[i+1:]
	}
	return strings.HasSuffix(label, "-security") ||
		archive == "security" ||
		strings.HasSuffix(archive, "-security")
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
		inst, ok := parseAptInstLine(string(line))
		if !ok {
			continue
		}
		name := inst.Name

		dpkgOut, err := runner.RunBytes("dpkg-query", "-W", "-f="+dpkgQueryFormat, name)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		info := parseDpkgQuery(name, dpkgOut)

		pkg := OsPackage{Name: name,
			Current_version:   info.Version,
			Candidate_version: inst.Candidate,
			Priority:          info.Priority,
			Section:           info.Section,
			Security:          inst.Security,
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
