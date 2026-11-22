package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func withTrackerURL(t *testing.T, url string) {
	t.Helper()
	orig := trackerAPIURL
	trackerAPIURL = url
	vulns = nil
	t.Cleanup(func() {
		trackerAPIURL = orig
		vulns = nil
	})
}

func TestPreloadVulnsConnectionRefused(t *testing.T) {
	// Grab a free port and close it so the connection is refused.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	l.Close()
	withTrackerURL(t, "http://"+addr+"/json")

	preloadVulns() // used to panic on a nil resp
	if len(vulns) != 0 {
		t.Errorf("got %d vulns, want 0", len(vulns))
	}
}

func TestPreloadVulnsServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `[{"name":"AVG-1","status":"Vulnerable","packages":["vim"]}]`, http.StatusInternalServerError)
	}))
	defer server.Close()
	withTrackerURL(t, server.URL)

	preloadVulns()
	if len(vulns) != 0 {
		t.Errorf("got %d vulns from a 500 response, want 0", len(vulns))
	}
}

func TestPreloadVulnsOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"name":"AVG-1","status":"Vulnerable","packages":["vim"]}]`))
	}))
	defer server.Close()
	withTrackerURL(t, server.URL)

	preloadVulns()
	if len(vulns) != 1 || vulns[0].Package[0] != "vim" {
		t.Errorf("got %+v, want one vim vuln", vulns)
	}
}

func TestArchInstalledInventory(t *testing.T) {
	Convey("pacman -Q output becomes an inventory", t, func() {
		r := new(MockRunner)
		r.On("Run", "pacman", []string{"-Q"}).Return("glibc 2.39-1\nlinux 6.8.2.arch1-1\nbad\n", nil)
		runner = r
		pkgs := ArchPackageManager{}.BuildInstalledInventory()
		So(pkgs, ShouldResemble, []InstalledPackage{
			{Name: "glibc", Version: "2.39-1"},
			{Name: "linux", Version: "6.8.2.arch1-1"},
		})
	})
}

func TestAgentDataInventoryJSON(t *testing.T) {
	Convey("installed_inventory is additive and omitted when empty", t, func() {
		s, _ := AgentData{InstalledInventory: []InstalledPackage{{Name: "a", Version: "1"}}}.ToJson()
		So(s, ShouldContainSubstring, `"installed_inventory":[{"name":"a","version":"1"}]`)
		s, _ = AgentData{}.ToJson()
		So(s, ShouldNotContainSubstring, "installed_inventory")
		So(s, ShouldContainSubstring, `"installed_packages"`)
	})
}
