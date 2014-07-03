package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestApiCheckIn(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, "foobar")
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	server.URL = "http://10.0.2.2:5000/api/v1/jobs?uid=UUID&api_key=d4b6c0ebf64456b1bec50cc679b146ed77b88195d681b96a902d15299c1ed51a"
	defer server.Close()

	//resp, _ := http.Get(server.URL)
	//body, _ := ioutil.ReadAll(resp.Body)

	api = SyswardApi{httpClient: &http.Client{}}

	Convey("Get a list of jobs", t, func() {
		api.GetJobs()
	})

}
