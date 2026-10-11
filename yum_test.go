package main

import (
	"errors"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// fakeYumCommands swaps yumCommandOutput for a lookup keyed on the full
// command line and restores the real implementation afterwards.
func fakeYumCommands(outputs map[string]string) func() {
	orig := yumCommandOutput
	yumCommandOutput = func(name string, args ...string) ([]byte, error) {
		key := strings.Join(append([]string{name}, args...), " ")
		out, ok := outputs[key]
		if !ok {
			return nil, errors.New("unexpected command: " + key)
		}
		return []byte(out), nil
	}
	return func() { yumCommandOutput = orig }
}

const evrQf = "%{EPOCHNUM}:%{VERSION}-%{RELEASE}\n"

func TestCentosBuildPackageList(t *testing.T) {
	Convey("Given dnf reports security and regular updates", t, func() {
		restore := fakeYumCommands(map[string]string{
			"dnf list updates --security": `Last metadata expiration check: 0:01:02 ago.
Available Upgrades
kernel.x86_64                 5.14.0-427.13.1.el9_4     baseos
openssl-libs.x86_64           1:3.0.7-27.el9            baseos
`,
			"dnf list updates": `Available Upgrades
kernel.x86_64                 5.14.0-427.13.1.el9_4     baseos
openssl-libs.x86_64           1:3.0.7-27.el9            baseos
tzdata.noarch                 2024a-1.el9               baseos
`,
			"rpm -q kernel.x86_64":                          "kernel-5.14.0-362.24.1.el9_3.x86_64\nkernel-5.14.0-70.13.1.el9_0.x86_64\nkernel-5.14.0-362.8.1.el9_3.x86_64\n",
			"rpm -q --qf " + evrQf + " kernel.x86_64":       "0:5.14.0-362.24.1.el9_3\n0:5.14.0-70.13.1.el9_0\n0:5.14.0-362.8.1.el9_3\n",
			"rpm -q openssl-libs.x86_64":                    "openssl-libs-3.0.7-24.el9.x86_64\n",
			"rpm -q --qf " + evrQf + " openssl-libs.x86_64": "1:3.0.7-24.el9\n",
			"rpm -q tzdata.noarch":                          "tzdata-2023c-1.el9.noarch\n",
			"rpm -q --qf " + evrQf + " tzdata.noarch":       "0:2023c-1.el9\n",
		})
		defer restore()

		pkgs := CentosPackageManager{}.BuildPackageList()

		Convey("Each package appears once (regression: the --security output was parsed twice)", func() {
			So(len(pkgs), ShouldEqual, 3)
			seen := map[string]int{}
			for _, p := range pkgs {
				seen[p.Name]++
			}
			So(seen, ShouldResemble, map[string]int{"kernel.x86_64": 1, "openssl-libs.x86_64": 1, "tzdata.noarch": 1})
		})

		Convey("Existing fields keep their meaning", func() {
			So(pkgs[0].Name, ShouldEqual, "kernel.x86_64")
			So(pkgs[0].Current_version, ShouldEqual, "kernel-5.14.0-362.24.1.el9_3.x86_64\nkernel-5.14.0-70.13.1.el9_0.x86_64\nkernel-5.14.0-362.8.1.el9_3.x86_64")
			So(pkgs[0].Candidate_version, ShouldEqual, "5.14.0-427.13.1.el9_4")
			So(pkgs[0].Section, ShouldEqual, "baseos")
			So(pkgs[0].Priority, ShouldEqual, "N/A")
			So(pkgs[0].Security, ShouldBeTrue)
			So(pkgs[1].Security, ShouldBeTrue)
			So(pkgs[2].Security, ShouldBeFalse)
		})

		Convey("New fields: arch from name.arch, highest EVR, empty source", func() {
			So(pkgs[0].Arch, ShouldEqual, "x86_64")
			So(pkgs[0].InstalledEVR, ShouldEqual, "0:5.14.0-362.24.1.el9_3")
			So(pkgs[0].Source, ShouldEqual, "")
			So(pkgs[1].InstalledEVR, ShouldEqual, "1:3.0.7-24.el9")
			So(pkgs[2].Arch, ShouldEqual, "noarch")
			So(pkgs[2].InstalledEVR, ShouldEqual, "0:2023c-1.el9")
		})
	})

	Convey("Given ForceYum and a failing EVR query", t, func() {
		restore := fakeYumCommands(map[string]string{
			"yum list updates --security": "",
			"yum list updates":            "bash.x86_64   5.1.8-9.el9   baseos\n",
			"rpm -q bash.x86_64":          "bash-5.1.8-6.el9.x86_64\n",
		})
		defer restore()

		pkgs := CentosPackageManager{ForceYum: true}.BuildPackageList()
		So(len(pkgs), ShouldEqual, 1)
		So(pkgs[0].Current_version, ShouldEqual, "bash-5.1.8-6.el9.x86_64")
		So(pkgs[0].InstalledEVR, ShouldEqual, "")
		So(pkgs[0].Arch, ShouldEqual, "x86_64")
	})
}

func TestRpmArch(t *testing.T) {
	Convey("rpmArch takes the last dot-separated part", t, func() {
		So(rpmArch("kernel.x86_64"), ShouldEqual, "x86_64")
		So(rpmArch("python3.11-libs.aarch64"), ShouldEqual, "aarch64")
		So(rpmArch("noarchname"), ShouldEqual, "")
		So(rpmArch("trailing."), ShouldEqual, "")
	})
}

func TestHighestEVR(t *testing.T) {
	Convey("highestEVR picks the newest regardless of order", t, func() {
		So(highestEVR("0:4.18.0-553.el8_10\n0:4.18.0-80.el8\n"), ShouldEqual, "0:4.18.0-553.el8_10")
		So(highestEVR("0:4.18.0-80.el8\n0:4.18.0-553.el8_10\n"), ShouldEqual, "0:4.18.0-553.el8_10")
		So(highestEVR("0:2.0-1\n1:1.0-1\n"), ShouldEqual, "1:1.0-1")
		So(highestEVR("\n"), ShouldEqual, "")
		So(highestEVR("0:1.0-1"), ShouldEqual, "0:1.0-1")
	})
}

func TestRpmvercmp(t *testing.T) {
	// Cases from rpm's own test suite (tests/rpmvercmp.at).
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0", "1.0", 0},
		{"1.0", "2.0", -1},
		{"2.0", "1.0", 1},
		{"2.0.1", "2.0.1", 0},
		{"2.0", "2.0.1", -1},
		{"2.0.1a", "2.0.1", 1},
		{"5.5p1", "5.5p2", -1},
		{"5.5p10", "5.5p1", 1},
		{"10xyz", "10.1xyz", -1},
		{"xyz10", "xyz10.1", -1},
		{"xyz.4", "8", -1},
		{"5.5p1", "5.5.p1", 0},
		{"10b2", "10a1", 1},
		{"1.0010", "1.9", 1},
		{"1.05", "1.5", 0},
		{"1.0", "1", 1},
		{"2.50", "2.5", 1},
		{"fc4", "fc.4", 0},
		{"FC5", "fc4", -1},
		{"2a", "2.0", -1},
		{"1.0", "1.fc4", 1},
		{"3.0.0_fc", "3.0.0.fc", 0},
		{"1++", "1_", 0},
		{"+", "_", 0},
		{"1.0~rc1", "1.0", -1},
		{"1.0~rc1", "1.0~rc2", -1},
		{"1.0~rc1~git123", "1.0~rc1", -1},
		{"1.0^", "1.0", 1},
		{"1.0^git1", "1.0", 1},
		{"1.0^git1", "1.01", -1},
		{"1.0^20160101", "1.0.1", -1},
		{"1.0^20160101^git1", "1.0^20160101", 1},
		{"1.0~rc1^git1", "1.0~rc1", 1},
		{"1.0^git1~pre", "1.0^git1", -1},
	}
	Convey("rpmvercmp matches rpm's reference results", t, func() {
		for _, c := range cases {
			So(c.a+" vs "+c.b+" = "+itoa(rpmvercmp(c.a, c.b)), ShouldEqual, c.a+" vs "+c.b+" = "+itoa(c.want))
			So(c.b+" vs "+c.a+" = "+itoa(rpmvercmp(c.b, c.a)), ShouldEqual, c.b+" vs "+c.a+" = "+itoa(-c.want))
		}
	})
}

func itoa(i int) string {
	switch {
	case i < 0:
		return "-1"
	case i > 0:
		return "1"
	}
	return "0"
}
