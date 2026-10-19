package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPackagesThatNeedUpdates(t *testing.T) {
	packageManager = DebianPackageManager{}
	Convey("Given pending updates", t, func() {

		Convey("There should be a list of packages available for update", func() {
			mockValue := `
Inst apport [2.20.11-0ubuntu82.4] (2.20.11-0ubuntu82.5 Ubuntu:22.04/jammy-updates [all])
`
			dpkMock := "2.20.11-0ubuntu82.4\tapport\tall\toptional\tutils\n"
			r := new(MockRunner)
			r.On("RunBytes", "apt-get", []string{"-s", "upgrade"}).Return(mockValue, nil)
			r.On("RunBytes", "dpkg-query", []string{"-W", "-f=" + dpkgQueryFormat, "apport"}).Return(dpkMock, nil)
			runner = r
			osPackages := packageManager.BuildPackageList()
			So(osPackages[0].Name, ShouldEqual, "apport")
			So(osPackages[0].Security, ShouldEqual, false)
			So(osPackages[0].Current_version, ShouldEqual, "2.20.11-0ubuntu82.4")
			So(osPackages[0].Candidate_version, ShouldEqual, "2.20.11-0ubuntu82.5")
			So(osPackages[0].Priority, ShouldEqual, "optional")
			So(osPackages[0].Section, ShouldEqual, "utils")
			So(osPackages[0].Source, ShouldEqual, "apport")
			So(osPackages[0].Arch, ShouldEqual, "all")
			So(osPackages[0].InstalledEVR, ShouldEqual, "2.20.11-0ubuntu82.4")
			r.Mock.AssertExpectations(t)
		})
	})
}

func TestPackageListSourceArchAndEVR(t *testing.T) {
	packageManager = DebianPackageManager{}
	Convey("Given pending updates whose binary and source names differ", t, func() {
		mockValue := `
Inst libssl3 [3.0.2-0ubuntu1.15] (3.0.2-0ubuntu1.18 Debian-Security:12/stable-security [amd64])
Inst libc6:i386 [2.35-0ubuntu3.6] (2.35-0ubuntu3.8 Ubuntu:22.04/jammy-updates [i386])
Inst tzdata [1:2024a-0ubuntu0.22.04] (1:2024b-0ubuntu0.22.04 Ubuntu:22.04/jammy-updates [all])
`
		r := new(MockRunner)
		r.On("RunBytes", "apt-get", []string{"-s", "upgrade"}).Return(mockValue, nil)
		r.On("RunBytes", "dpkg-query", []string{"-W", "-f=" + dpkgQueryFormat, "libssl3"}).
			Return("3.0.2-0ubuntu1.15\topenssl\tamd64\toptional\tlibs\n", nil)
		r.On("RunBytes", "dpkg-query", []string{"-W", "-f=" + dpkgQueryFormat, "libc6:i386"}).
			Return("2.35-0ubuntu3.6\tglibc\ti386\toptional\tlibs\n", nil)
		// older dpkg without source:Package support yields an empty source
		r.On("RunBytes", "dpkg-query", []string{"-W", "-f=" + dpkgQueryFormat, "tzdata"}).
			Return("1:2024a-0ubuntu0.22.04\t\tall\trequired\tlocalization\n", nil)
		runner = r

		pkgs := packageManager.BuildPackageList()
		So(len(pkgs), ShouldEqual, 3)

		So(pkgs[0].Name, ShouldEqual, "libssl3")
		So(pkgs[0].Source, ShouldEqual, "openssl")
		So(pkgs[0].Arch, ShouldEqual, "amd64")
		So(pkgs[0].InstalledEVR, ShouldEqual, "3.0.2-0ubuntu1.15")
		So(pkgs[0].Current_version, ShouldEqual, "3.0.2-0ubuntu1.15")
		So(pkgs[0].Security, ShouldBeTrue)

		So(pkgs[1].Name, ShouldEqual, "libc6:i386")
		So(pkgs[1].Source, ShouldEqual, "glibc")
		So(pkgs[1].Arch, ShouldEqual, "i386")

		So(pkgs[2].Source, ShouldEqual, "tzdata")
		So(pkgs[2].InstalledEVR, ShouldEqual, "1:2024a-0ubuntu0.22.04")
		So(pkgs[2].Current_version, ShouldEqual, "1:2024a-0ubuntu0.22.04")
		So(pkgs[2].Priority, ShouldEqual, "required")
		So(pkgs[2].Section, ShouldEqual, "localization")
		r.Mock.AssertExpectations(t)
	})
}

func TestParseDpkgQuery(t *testing.T) {
	Convey("Given dpkg-query output", t, func() {
		Convey("Multi-arch duplicates use the first line", func() {
			info := parseDpkgQuery("libc6", []byte("2.35-0ubuntu3.6\tglibc\tamd64\trequired\tlibs\n2.35-0ubuntu3.6\tglibc\ti386\trequired\tlibs\n"))
			So(info.Version, ShouldEqual, "2.35-0ubuntu3.6")
			So(info.Arch, ShouldEqual, "amd64")
		})
		Convey("Empty source falls back to the name without arch qualifier", func() {
			info := parseDpkgQuery("libfoo:i386", []byte("1.0-1\t\ti386\toptional\tlibs\n"))
			So(info.Source, ShouldEqual, "libfoo")
		})
		Convey("Short output does not panic", func() {
			info := parseDpkgQuery("foo", []byte("1.0-1"))
			So(info.Version, ShouldEqual, "1.0-1")
			So(info.Source, ShouldEqual, "foo")
			So(info.Arch, ShouldEqual, "")
		})
	})
}

func TestChangeLog(t *testing.T) {
	packageManager = DebianPackageManager{}
	Convey("Changelong gets 64bit encoded", t, func() {
		r := new(MockRunner)
		r.On("Run", "apt-get", []string{"changelog", "apt"}).Return("foobar", nil)
		runner = r
		So(packageManager.GetChangelog("apt"), ShouldEqual, "Zm9vYmFy")
		r.Mock.AssertExpectations(t)
	})
}

func TestPackageUpdates(t *testing.T) {
	packageManager = DebianPackageManager{}
	Convey("Given a package name", t, func() {

		Convey("The package should be upgraded", func() {
			r := new(MockRunner)
			r.On("Run", "apt-get", []string{
				"install",
				"-y",
				"-o",
				fmt.Sprintf("Dpkg::Options::=--force-confdef"),
				"-o",
				fmt.Sprintf("Dpkg::Options::=--force-confold"),
				"apt"}).Return("", nil)
			runner = r
			err := packageManager.UpdatePackage("apt")
			So(err, ShouldBeNil)
			r.Mock.AssertExpectations(t)
		})

		Convey("The package should not upgrade if held", func() {
			r := new(MockRunner)
			r.On("Run", "apt-get", []string{
				"install",
				"-y",
				"-o",
				fmt.Sprintf("Dpkg::Options::=--force-confdef"),
				"-o",
				fmt.Sprintf("Dpkg::Options::=--force-confold"),
				"apt"}).Return("", errors.New("fail"))
			runner = r
			err := packageManager.UpdatePackage("apt")
			So(err, ShouldNotBeNil)
			r.Mock.AssertExpectations(t)
		})
	})

}

func TestPackageHolding(t *testing.T) {
	packageManager = DebianPackageManager{}
	Convey("Given holding a package", t, func() {

		Convey("The package should be held", func() {
			r := new(MockRunner)
			r.On("Run", "apt-mark", []string{"hold", "apt"}).Return("", nil)
			runner = r
			err := packageManager.HoldPackage("apt")
			So(err, ShouldBeNil)
			r.Mock.AssertExpectations(t)
		})
	})

	Convey("Given unholding a package", t, func() {

		Convey("The package should be unheld", func() {
			r := new(MockRunner)
			r.On("Run", "apt-mark", []string{"unhold", "apt"}).Return("", nil)
			runner = r
			err := packageManager.UnholdPackage("apt")
			So(err, ShouldBeNil)
			r.Mock.AssertExpectations(t)
		})
	})

}

func TestSourceList(t *testing.T) {
	packageManager = DebianPackageManager{}
	Convey("Given /etc/apt/sources.list exists", t, func() {

		Convey("There should be a source list", func() {
			packageList := []string{"deb http://us.archive.ubuntu.com/ubuntu/ trusty main restricted",
				"deb-src http://us.archive.ubuntu.com/ubuntu/ trusty main restricted"}
			r := new(MockRunner)
			r.On("Run", "grep", []string{"-h", "^deb", "/etc/apt/sources.list", "/etc/apt/sources.list.d/*"}).Return(strings.Join(packageList, "\n"), nil)
			runner = r
			sourcesList := packageManager.GetSourcesList()
			src_one := sourcesList[0]
			src_two := sourcesList[1]
			So(src_one.Url, ShouldEqual, "http://us.archive.ubuntu.com/ubuntu/")
			So(src_one.Src, ShouldBeFalse)
			So(src_two.Src, ShouldBeTrue)
			r.Mock.AssertExpectations(t)
		})
	})

}

func TestInstalledPackages(t *testing.T) {
	packageManager = DebianPackageManager{}
	Convey("Given I want to view all installed packages", t, func() {
		Convey("It returns a list of all installed packages", func() {
			r := new(MockRunner)
			r.On("Run", "dpkg", []string{"--get-selections"}).Return("apt\u0009installed", nil)
			runner = r
			packages := packageManager.BuildInstalledPackageList()
			So(packages[0], ShouldEqual, "apt")
			So(len(packages), ShouldEqual, 1)
			r.Mock.AssertExpectations(t)
		})
	})

}

func TestUpdatingThePackageList(t *testing.T) {
	packageManager = DebianPackageManager{}
	Convey("Given I want to have the latest source list", t, func() {

		Convey("apt-update gets run", func() {
			r := new(MockRunner)
			r.On("Run", "apt-get", []string{"update"}).Return("", nil)
			runner = r
			err := packageManager.UpdatePackageLists()
			So(err, ShouldBeNil)
			r.Mock.AssertExpectations(t)
		})
	})

}

//func TestUpdateCounts(t *testing.T) {
//	packageManager = DebianPackageManager{}
//	Convey("Given there are security and regular updates", t, func() {
//
//		Convey("The number of security and regular updates is > 0", func() {
//			//r := new(MockRunner)
//			//r.On("Run", "/usr/lib/update-notifier/apt-check", []string{}).Return("1;2", nil)
//			//runner = r
//			updates := packageManager.UpdateCounts()
//			So(updates.Regular, ShouldEqual, 1)
//			So(updates.Security, ShouldEqual, 2)
//			//r.Mock.AssertExpectations(t)
//		})
//
//		Convey("There are no security updates", func() {
//			r := new(MockRunner)
//			r.On("Run", "/usr/lib/update-notifier/apt-check", []string{}).Return("2;0", nil)
//			runner = r
//			updates := packageManager.UpdateCounts()
//			So(updates.Regular, ShouldEqual, 2)
//			So(updates.Security, ShouldEqual, 0)
//			r.Mock.AssertExpectations(t)
//		})
//
//	})
//}

// Lines in the shape `apt-get -s upgrade` prints them: the parenthesised part
// is "candidate Label:Version/Archive[, Label:Version/Archive...] [arch]".
// The "captured" cases are verbatim from Ubuntu 22.04/24.04 hosts; the rest
// follow apt's RelStr format for releases those hosts don't carry.
func TestParseAptInstLine(t *testing.T) {
	cases := []struct {
		name      string
		line      string
		pkg       string
		candidate string
		security  bool
	}{
		// Captured from real hosts (apt 2.4.12 on 22.04, apt 2.7 on 24.04). Ubuntu
		// publishes security fixes to both pockets and -updates is listed first.
		{"captured 22.04 updates+security", "Inst libssl3 [3.0.2-0ubuntu1.30] (3.0.2-0ubuntu1.30 Ubuntu:22.04/jammy-updates, Ubuntu:22.04/jammy-security [amd64])", "libssl3", "3.0.2-0ubuntu1.30", true},
		{"captured 24.04 updates+security", "Inst libssl3t64 [3.0.13-0ubuntu3.16] (3.0.13-0ubuntu3.16 Ubuntu:24.04/noble-updates, Ubuntu:24.04/noble-security [amd64])", "libssl3t64", "3.0.13-0ubuntu3.16", true},
		{"captured 22.04 updates only", "Inst apt [2.4.12] (2.4.14 Ubuntu:22.04/jammy-updates [amd64])", "apt", "2.4.14", false},
		{"captured 24.04 updates only", "Inst login [1:4.13+dfsg1-4ubuntu3] (1:4.13+dfsg1-4ubuntu3.2 Ubuntu:24.04/noble-updates [amd64])", "login", "1:4.13+dfsg1-4ubuntu3.2", false},
		{"captured trailing broken-deps brackets", "Inst libext2fs2 [1.46.5-2ubuntu1.1] (1.46.5-2ubuntu1.2 Ubuntu:22.04/jammy-updates [amd64]) [e2fsprogs:amd64 on libext2fs2:amd64] [e2fsprogs:amd64 ]", "libext2fs2", "1.46.5-2ubuntu1.2", false},
		{"captured trailing empty brackets", "Inst initramfs-tools [0.140ubuntu13.4] (0.140ubuntu13.5 Ubuntu:22.04/jammy-updates [all]) []", "initramfs-tools", "0.140ubuntu13.5", false},
		{"captured third-party repo", "Inst zabbix-agent2 [1:7.4.6-1+ubuntu22.04] (1:7.4.15-1+ubuntu22.04 zabbix:repo.zabbix.com [amd64])", "zabbix-agent2", "1:7.4.15-1+ubuntu22.04", false},

		{"ubuntu 20.04 focal-security", "Inst libssl1.1 [1.1.1f-1ubuntu2.20] (1.1.1f-1ubuntu2.22 Ubuntu:20.04/focal-security [amd64])", "libssl1.1", "1.1.1f-1ubuntu2.22", true},
		{"ubuntu 22.04 jammy-security", "Inst libssl3 [3.0.2-0ubuntu1.12] (3.0.2-0ubuntu1.15 Ubuntu:22.04/jammy-security [amd64])", "libssl3", "3.0.2-0ubuntu1.15", true},
		{"ubuntu 22.04 updates and security", "Inst libc6 [2.35-0ubuntu3.5] (2.35-0ubuntu3.6 Ubuntu:22.04/jammy-updates, Ubuntu:22.04/jammy-security [arm64])", "libc6", "2.35-0ubuntu3.6", true},
		{"ubuntu 22.04 security listed first", "Inst libc6 [2.35-0ubuntu3.5] (2.35-0ubuntu3.6 Ubuntu:22.04/jammy-security, Ubuntu:22.04/jammy-updates [arm64])", "libc6", "2.35-0ubuntu3.6", true},
		{"ubuntu 24.04 noble-security", "Inst libgnutls30t64 [3.8.3-1.1ubuntu3] (3.8.3-1.1ubuntu3.2 Ubuntu:24.04/noble-updates, Ubuntu:24.04/noble-security [amd64])", "libgnutls30t64", "3.8.3-1.1ubuntu3.2", true},
		{"ubuntu pro esm", "Inst libcurl4 [7.68.0-1ubuntu2.21] (7.68.0-1ubuntu2.22+esm1 UbuntuESM:20.04/focal-infra-security [amd64])", "libcurl4", "7.68.0-1ubuntu2.22+esm1", true},
		{"debian 12 Debian-Security", "Inst libssl3 [3.0.11-1~deb12u1] (3.0.11-1~deb12u2 Debian-Security:12/stable-security [amd64])", "libssl3", "3.0.11-1~deb12u2", true},
		{"debian 12 point release and security", "Inst libc6 [2.36-9+deb12u3] (2.36-9+deb12u4 Debian:12.5/stable, Debian-Security:12/stable-security [amd64])", "libc6", "2.36-9+deb12u4", true},
		{"debian 12 bookworm-security suite", "Inst libssl3 [3.0.11-1~deb12u1] (3.0.11-1~deb12u2 Debian:12/bookworm-security [amd64])", "libssl3", "3.0.11-1~deb12u2", true},
		{"debian 11 bullseye-security", "Inst libssl1.1 [1.1.1n-0+deb11u4] (1.1.1n-0+deb11u5 Debian-Security:11/oldstable-security [amd64])", "libssl1.1", "1.1.1n-0+deb11u5", true},
		{"debian 10 label only", "Inst libssl1.1 [1.1.1d-0+deb10u7] (1.1.1d-0+deb10u8 Debian-Security:10/oldoldstable [amd64])", "libssl1.1", "1.1.1d-0+deb10u8", true},
		{"lowercase debian-security label", "Inst openssl [3.0.11-1~deb12u1] (3.0.11-1~deb12u2 debian-security:12/stable-security [arm64])", "openssl", "3.0.11-1~deb12u2", true},
		{"new install without current version", "Inst linux-image-6.8.0-45-generic (6.8.0-45.45 Ubuntu:24.04/noble-security [amd64])", "linux-image-6.8.0-45-generic", "6.8.0-45.45", true},
		{"epoch version", "Inst tzdata [2024a-0ubuntu0.22.04] (2024a-0ubuntu0.22.04.1 Ubuntu:22.04/jammy-updates [all])", "tzdata", "2024a-0ubuntu0.22.04.1", false},

		{"ubuntu jammy-updates", "Inst apport [2.20.11-0ubuntu82.4] (2.20.11-0ubuntu82.5 Ubuntu:22.04/jammy-updates [all])", "apport", "2.20.11-0ubuntu82.5", false},
		{"ubuntu noble-updates", "Inst systemd [255.4-1ubuntu8] (255.4-1ubuntu8.4 Ubuntu:24.04/noble-updates [amd64])", "systemd", "255.4-1ubuntu8.4", false},
		{"ubuntu backports", "Inst cockpit [264-1] (300-1~ubuntu22.04.1 Ubuntu:22.04/jammy-backports [all])", "cockpit", "300-1~ubuntu22.04.1", false},
		{"debian point release", "Inst base-files [12.4+deb12u4] (12.4+deb12u5 Debian:12.5/stable [amd64])", "base-files", "12.4+deb12u5", false},
		{"debian stable-updates", "Inst tzdata [2023c-5] (2024a-0+deb12u1 Debian:stable-updates [all])", "tzdata", "2024a-0+deb12u1", false},
		{"debian backports", "Inst golang-1.21 [1.21.3-1~bpo12+1] (1.21.6-1~bpo12+1 Debian Backports:bookworm-backports [all])", "golang-1.21", "1.21.6-1~bpo12+1", false},
		{"ppa", "Inst php8.3-cli [8.3.10-1+ubuntu22.04.1+deb.sury.org+1] (8.3.11-1+ubuntu22.04.1+deb.sury.org+1 LP-PPA-ondrej-php:22.04/jammy [amd64])", "php8.3-cli", "8.3.11-1+ubuntu22.04.1+deb.sury.org+1", false},
		{"package name contains security", "Inst wazuh-security-agent [4.7.0-1] (4.7.1-1 wazuh:stable [amd64])", "wazuh-security-agent", "4.7.1-1", false},
		{"version contains security", "Inst libfoo [1.0-1] (1.0-1security1 Ubuntu:22.04/jammy-updates [amd64])", "libfoo", "1.0-1security1", false},
		{"arch-qualified name", "Inst libc6:i386 [2.35-0ubuntu3.5] (2.35-0ubuntu3.6 Ubuntu:22.04/jammy-updates [i386])", "libc6:i386", "2.35-0ubuntu3.6", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := parseAptInstLine(c.line)
			if !ok {
				t.Fatalf("line not parsed: %q", c.line)
			}
			if got.Name != c.pkg || got.Candidate != c.candidate || got.Security != c.security {
				t.Errorf("got %+v, want {Name:%s Candidate:%s Security:%v}", got, c.pkg, c.candidate, c.security)
			}
		})
	}
}

func TestParseAptInstLineIgnoresOtherLines(t *testing.T) {
	for _, line := range []string{
		"",
		"Reading package lists...",
		"The following packages will be upgraded:",
		"   libssl3 openssl",
		"2 upgraded, 0 newly installed, 0 to remove and 0 not upgraded.",
		"Conf libssl3 (3.0.2-0ubuntu1.15 Ubuntu:22.04/jammy-security [amd64])",
		"Remv foo [1.0-1]",
	} {
		if got, ok := parseAptInstLine(line); ok {
			t.Errorf("%q parsed as %+v", line, got)
		}
	}
}

func TestUpdateCountsUbuntuSecurity(t *testing.T) {
	packageManager = DebianPackageManager{}
	mockValue := `Reading package lists...
Building dependency tree...
Reading state information...
Calculating upgrade...
The following packages will be upgraded:
  apport libssl3 openssl
3 upgraded, 0 newly installed, 0 to remove and 0 not upgraded.
Inst apport [2.20.11-0ubuntu82.4] (2.20.11-0ubuntu82.5 Ubuntu:22.04/jammy-updates [all])
Inst libssl3 [3.0.2-0ubuntu1.12] (3.0.2-0ubuntu1.15 Ubuntu:22.04/jammy-updates, Ubuntu:22.04/jammy-security [amd64])
Inst openssl [3.0.2-0ubuntu1.12] (3.0.2-0ubuntu1.15 Ubuntu:22.04/jammy-security [amd64])
Conf apport (2.20.11-0ubuntu82.5 Ubuntu:22.04/jammy-updates [all])
Conf libssl3 (3.0.2-0ubuntu1.15 Ubuntu:22.04/jammy-updates, Ubuntu:22.04/jammy-security [amd64])
Conf openssl (3.0.2-0ubuntu1.15 Ubuntu:22.04/jammy-security [amd64])
`
	r := new(MockRunner)
	r.On("RunBytes", "apt-get", []string{"-s", "upgrade"}).Return(mockValue, nil)
	for _, name := range []string{"apport", "libssl3", "openssl"} {
		r.On("RunBytes", "dpkg-query", []string{"-W", "-f=" + dpkgQueryFormat, name}).
			Return("1.0\t"+name+"\tamd64\toptional\tlibs\n", nil)
	}
	runner = r
	updates := packageManager.UpdateCounts()
	if updates.Security != 2 || updates.Regular != 1 {
		t.Errorf("got %+v, want 2 security and 1 regular", updates)
	}
	r.Mock.AssertExpectations(t)
}
