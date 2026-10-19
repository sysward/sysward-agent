package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
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
