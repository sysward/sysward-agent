package main

import (
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

// yumCommandOutput runs a command and returns its stdout only (stderr carries
// dnf metadata chatter that must not be parsed as package lines). It is a
// variable so tests can fake command output.
var yumCommandOutput = func(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

func (pm CentosPackageManager) BuildPackageList() []OsPackage {
	// build list of security updates first
	pkgManager := "dnf"
	if pm.ForceYum {
		pkgManager = "yum"
	}
	out, err := yumCommandOutput(pkgManager, "list", "updates", "--security")
	if err != nil {
		panic(err)
	}

	// Parse the command output to extract update information

	security := map[string]struct{}{}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		security[fields[0]] = struct{}{}
	}

	// Separate output from the security run: reusing one buffer for both
	// commands parsed every security package twice.
	out, err = yumCommandOutput(pkgManager, "list", "updates")
	if err != nil {
		panic(err)
	}

	// Parse the command output to extract update information
	var updates []OsPackage
	lines = strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		// Use rpm command to get the installed version
		rpmOut, err := yumCommandOutput("rpm", "-q", fields[0])
		if err != nil {
			continue // Skip if we can't get the installed version
		}
		installedVersion := strings.TrimSpace(string(rpmOut))

		installedEVR := ""
		evrOut, err := yumCommandOutput("rpm", "-q", "--qf", "%{EPOCHNUM}:%{VERSION}-%{RELEASE}\n", fields[0])
		if err == nil {
			installedEVR = highestEVR(string(evrOut))
		}

		_, isSecurity := security[fields[0]]

		// Create an Update struct and append it to the list
		updates = append(updates, OsPackage{
			Name:              fields[0],
			Current_version:   installedVersion,
			Candidate_version: fields[1],
			// The Priority and Section information are not available directly from dnf list updates command
			// If they are available from another source, you will need to adjust this part
			Priority:     "N/A",
			Section:      fields[2],
			Security:     isSecurity,
			Arch:         rpmArch(fields[0]),
			InstalledEVR: installedEVR,
		})
	}

	return updates
}

// rpmArch returns the arch suffix of a name.arch string ("" when absent).
func rpmArch(nameArch string) string {
	idx := strings.LastIndex(nameArch, ".")
	if idx < 0 || idx == len(nameArch)-1 {
		return ""
	}
	return nameArch[idx+1:]
}

// highestEVR picks the highest epoch:version-release from rpm --qf output.
// Several lines come back when more than one instance is installed (kernels).
// rpm prints them in rpmdb order, which is not guaranteed to be version order,
// so they are compared with rpmvercmp semantics rather than taking the last.
func highestEVR(out string) string {
	best := ""
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if best == "" || compareEVR(line, best) > 0 {
			best = line
		}
	}
	return best
}

// splitEVR splits "epoch:version-release" into its parts. A missing epoch is 0.
func splitEVR(evr string) (epoch, version, release string) {
	epoch = "0"
	if idx := strings.Index(evr, ":"); idx >= 0 {
		epoch = evr[:idx]
		evr = evr[idx+1:]
	}
	version = evr
	if idx := strings.LastIndex(evr, "-"); idx >= 0 {
		version = evr[:idx]
		release = evr[idx+1:]
	}
	if epoch == "" {
		epoch = "0"
	}
	return epoch, version, release
}

// compareEVR compares two epoch:version-release strings like rpm does.
func compareEVR(a, b string) int {
	ae, av, ar := splitEVR(a)
	be, bv, br := splitEVR(b)
	if c := rpmvercmp(ae, be); c != 0 {
		return c
	}
	if c := rpmvercmp(av, bv); c != 0 {
		return c
	}
	return rpmvercmp(ar, br)
}

func isASCIIDigit(c byte) bool { return c >= '0' && c <= '9' }
func isASCIIAlpha(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
func isASCIIAlnum(c byte) bool { return isASCIIDigit(c) || isASCIIAlpha(c) }

// rpmvercmp is a port of rpm's rpmvercmp (lib/rpmvercmp.c), including the
// '~' (sorts before anything) and '^' (sorts after the base, before more) rules.
func rpmvercmp(a, b string) int {
	if a == b {
		return 0
	}
	i, j := 0, 0
	for i < len(a) || j < len(b) {
		for i < len(a) && !isASCIIAlnum(a[i]) && a[i] != '~' && a[i] != '^' {
			i++
		}
		for j < len(b) && !isASCIIAlnum(b[j]) && b[j] != '~' && b[j] != '^' {
			j++
		}

		// tilde sorts before everything else, including the end of the string
		if (i < len(a) && a[i] == '~') || (j < len(b) && b[j] == '~') {
			if i >= len(a) || a[i] != '~' {
				return 1
			}
			if j >= len(b) || b[j] != '~' {
				return -1
			}
			i++
			j++
			continue
		}

		// caret sorts after the end of the string but before anything else
		if (i < len(a) && a[i] == '^') || (j < len(b) && b[j] == '^') {
			if i >= len(a) {
				return -1
			}
			if j >= len(b) {
				return 1
			}
			if a[i] != '^' {
				return 1
			}
			if b[j] != '^' {
				return -1
			}
			i++
			j++
			continue
		}

		if i >= len(a) || j >= len(b) {
			break
		}

		si, sj := i, j
		isNum := isASCIIDigit(a[i])
		if isNum {
			for i < len(a) && isASCIIDigit(a[i]) {
				i++
			}
			for j < len(b) && isASCIIDigit(b[j]) {
				j++
			}
		} else {
			for i < len(a) && isASCIIAlpha(a[i]) {
				i++
			}
			for j < len(b) && isASCIIAlpha(b[j]) {
				j++
			}
		}
		segA, segB := a[si:i], b[sj:j]

		// segments of different types: numeric is newer
		if segB == "" {
			if isNum {
				return 1
			}
			return -1
		}

		if isNum {
			segA = strings.TrimLeft(segA, "0")
			segB = strings.TrimLeft(segB, "0")
			if len(segA) != len(segB) {
				if len(segA) > len(segB) {
					return 1
				}
				return -1
			}
		}
		if c := strings.Compare(segA, segB); c != 0 {
			return c
		}
	}

	if i >= len(a) && j >= len(b) {
		return 0
	}
	if i >= len(a) {
		return -1
	}
	return 1
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
