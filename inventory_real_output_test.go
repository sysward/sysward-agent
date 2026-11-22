package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The files under testdata/inventory are the unmodified output of the exact
// inventory commands, captured from the distributions' official container
// images (2026-11-22). They pin the parsers to what real package managers
// print, including rpm 4.11 on CentOS 7 and the gpg-pubkey pseudo-packages.

func readInventoryFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "inventory", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestParseRpmInventoryOnRealOutput(t *testing.T) {
	evr := regexp.MustCompile(`^\d+:[^\s:-][^\s-]*-[^\s-]+$`)
	files, err := filepath.Glob(filepath.Join("testdata", "inventory", "rpm-*.txt"))
	if err != nil || len(files) < 7 {
		t.Fatalf("expected at least 7 rpm fixtures, got %d (%v)", len(files), err)
	}
	for _, f := range files {
		name := filepath.Base(f)
		raw := readInventoryFixture(t, name)
		want := 0
		for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
			if !strings.HasPrefix(line, "gpg-pubkey\t") {
				want++
			}
		}
		got := parseRpmInventory(raw)
		if len(got) != want || want < 50 {
			t.Errorf("%s: parsed %d packages, want %d", name, len(got), want)
		}
		seen := map[string]bool{}
		for _, p := range got {
			if p.Name == "" || p.Name == "gpg-pubkey" || strings.ContainsAny(p.Name, " \t") {
				t.Errorf("%s: bad name %q", name, p.Name)
			}
			if !evr.MatchString(p.Version) {
				t.Errorf("%s: %s version %q is not epoch:version-release", name, p.Name, p.Version)
			}
			if p.Arch == "" || p.Arch == "(none)" {
				t.Errorf("%s: %s has arch %q", name, p.Name, p.Arch)
			}
			seen[p.Name] = true
		}
		for _, must := range []string{"glibc", "bash"} {
			if !seen[must] {
				t.Errorf("%s: %s missing from the parsed inventory", name, must)
			}
		}
	}
}

func TestParseDpkgInventoryOnRealOutput(t *testing.T) {
	raw := readInventoryFixture(t, "dpkg-debian-12.txt")
	got := parseDpkgInventory(raw)
	want := 0
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		f := strings.Split(line, "\t")
		if len(f) == 5 && (strings.HasPrefix(f[4], "ii") || strings.HasPrefix(f[4], "hi")) {
			want++
		}
	}
	if len(got) != want || want < 50 {
		t.Fatalf("parsed %d packages, want %d", len(got), want)
	}
	bySource := map[string]string{}
	for _, p := range got {
		if p.Name == "" || p.Version == "" || p.Source == "" || p.Arch == "" {
			t.Errorf("incomplete entry %+v", p)
		}
		bySource[p.Name] = p.Source
	}
	// Debian's CVE data is keyed by source package: binaries must carry theirs.
	if bySource["libc6"] != "glibc" || bySource["bash"] != "bash" {
		t.Errorf("source packages wrong: libc6=%q bash=%q", bySource["libc6"], bySource["bash"])
	}
}

func TestParsePacmanInventoryOnRealOutput(t *testing.T) {
	raw := readInventoryFixture(t, "pacman-archlinux.txt")
	got := parsePacmanInventory(string(raw))
	want := len(strings.Split(strings.TrimSpace(string(raw)), "\n"))
	if len(got) != want || want < 50 {
		t.Fatalf("parsed %d packages, want %d", len(got), want)
	}
	for _, p := range got {
		if p.Name == "" || !strings.Contains(p.Version, "-") {
			t.Errorf("bad entry %+v", p)
		}
	}
}
