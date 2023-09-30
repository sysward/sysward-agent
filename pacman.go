package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sysward/sysward-agent/logging"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type ArchPackageManager struct{}

func (pm ArchPackageManager) UpdatePackage(pkg string) error {
	out, err := runner.Run("pacman",
		"-S",
		"--noconfirm",
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

func (pm ArchPackageManager) HoldPackage(p string) error {
	filename := "/etc/pacman.conf"
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Failed to open file: %s\n", err)
		return err
	}
	defer file.Close()

	var lines []string
	foundOptions := false
	ignorePkgLineUpdated := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)

		// Check if we are in the [options] block
		if line == "[options]" {
			foundOptions = true
		} else if foundOptions && (strings.HasPrefix(line, "[") || line == "") {
			// Exiting the [options] block
			foundOptions = false

			// If IgnorePkg line wasn't found, we add it
			if !ignorePkgLineUpdated {
				ignorePkgStr := "IgnorePkg = " + p
				lines = append(lines, ignorePkgStr)
			}
		} else if foundOptions && strings.HasPrefix(line, "IgnorePkg") {
			// Modify the IgnorePkg line to include the new packages
			existingPackages := strings.TrimPrefix(line, "IgnorePkg = ")
			pkgMap := map[string]bool{}
			for _, pkg := range strings.Fields(existingPackages) {
				pkgMap[pkg] = true
			}

			pkgMap[p] = true

			updatedPackages := make([]string, 0, len(pkgMap))
			for pkg := range pkgMap {
				updatedPackages = append(updatedPackages, pkg)
			}

			lines[len(lines)-1] = "IgnorePkg = " + strings.Join(updatedPackages, " ")
			ignorePkgLineUpdated = true
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Failed to read file: %s\n", err)
		return err
	}

	// Write the updated contents back to the file
	file, err = os.Create(filename)
	if err != nil {
		fmt.Printf("Failed to open file for writing: %s\n", err)
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	return writer.Flush()
}

// TODO: fix bug in unholding arch pkgs
func (pm ArchPackageManager) UnholdPackage(p string) error {
	filename := "/etc/pacman.conf"
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Failed to open file: %s\n", err)
		return err
	}
	defer file.Close()

	var lines []string
	foundOptions := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Check if we are in the [options] block
		if line == "[options]" {
			foundOptions = true
		} else if foundOptions && (strings.HasPrefix(line, "[") || line == "") {
			// Exiting the [options] block
			foundOptions = false
		} else if foundOptions && strings.HasPrefix(line, "IgnorePkg") {
			// Modify the IgnorePkg line to remove the specified packages
			existingPackages := strings.TrimPrefix(line, "IgnorePkg = ")
			pkgMap := map[string]bool{}
			for _, pkg := range strings.Fields(existingPackages) {
				pkgMap[pkg] = true
			}

			delete(pkgMap, p)

			// If no packages are left in the IgnorePkg line, skip adding this line to the new file content
			if len(pkgMap) == 0 {
				continue
			}

			updatedPackages := make([]string, 0, len(pkgMap))
			for pkg := range pkgMap {
				updatedPackages = append(updatedPackages, pkg)
			}

			line = "IgnorePkg = " + strings.Join(updatedPackages, " ")
		}

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Failed to read file: %s\n", err)
		return err
	}

	// Write the updated contents back to the file
	file, err = os.Create(filename)
	if err != nil {
		fmt.Printf("Failed to open file for writing: %s\n", err)
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	return writer.Flush()
}

func (pm ArchPackageManager) GetSourcesList() []Source {
	return []Source{}
}

// TODO: Not supported
func (pm ArchPackageManager) GetChangelog(package_name string) string {
	return ""
}

func (pm ArchPackageManager) BuildInstalledPackageList() []string {
	output, err := runner.Run("pacman", "-Q")
	if err != nil {
		fmt.Printf("Failed to fetch installed packages: %s\n", err)
		return []string{}
	}

	packages := []string{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 1 {
			continue
		}
		packages = append(packages, parts[0])
	}

	return packages
}

type Vulnerability struct {
	Name    string   `json:"name"`
	Status  string   `json:"status"`
	Package []string `json:"packages"`
}

const trackerAPIURL = "https://security.archlinux.org/json"

var vulns []Vulnerability

func preloadVulns() {
	resp, err := http.Get(trackerAPIURL)
	if err != nil {
		fmt.Println("Error fetching data:", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}

	if err := json.Unmarshal(data, &vulns); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}
	fmt.Println("Loaded vulns: ", len(vulns))
}

func isSecurityUpdate(p OsPackage) bool {
	for _, vuln := range vulns {
		for _, pkg := range vuln.Package {
			if pkg == p.Name && (vuln.Status == "open" || vuln.Status == "assigned") {
				fmt.Printf("The package %s has an open or assigned security issue: %s\n", p.Name, vuln.Name)
				return true
			}
		}
	}
	return false
}

func (pm ArchPackageManager) BuildPackageList() []OsPackage {
	preloadVulns()
	var packages []OsPackage

	output, err := runner.Run("pacman", "-Qu")
	if err != nil {
		fmt.Printf("Failed to fetch updates: %s\n", err)
		return []OsPackage{}
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			info := OsPackage{
				Name:              parts[0],
				Current_version:   parts[1],
				Candidate_version: parts[2],
				Priority:          "unknown",
				Section:           "unknown",
				Security:          false,
			}
			packages = append(packages, info)
		}
	}

	return packages
}

func (pm ArchPackageManager) UpdatePackageLists() error {
	_, err := runner.Run("pacman", "-Sy")
	return err
}

func (pm ArchPackageManager) UpdateCounts() Updates {
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
