package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCheckIn(t *testing.T) {
	expected := `{"packages":null,"system_updates":{"regular":0,"security":0},"operating_system":{"name":"","uid":"","version":"","interfaces":null,"hostname":"","cpu_information":{"name":"","count":0},"memory_information":{"total":""}},"sources":null,"installed_packages":null,"group":"","custom_hostname":"","hostname":"","reboot_required":false}`
	Convey("Checking in via the API", t, func(ctx C) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			body, _ := ioutil.ReadAll(r.Body)
			ctx.So(string(body), ShouldEqual, expected)
			w.WriteHeader(200)
		}
		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()
		api = SyswardApi{httpClient: http.Client{}}
		c := new(MockConfig)
		c.On("agentCheckinUrl").Return(server.URL)
		config = c

		api.CheckIn(AgentData{})
	})
}

func TestApiJobFailure(t *testing.T) {
	Convey("Job failure should send the job data to the server", t, func() {
		postedBody := ""
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			body, _ := ioutil.ReadAll(r.Body)
			postedBody = string(body)
			w.WriteHeader(200)
		}
		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()
		api = SyswardApi{httpClient: http.Client{}}
		c := new(MockConfig)
		c.On("fetchJobPostbackUrl").Return(server.URL)
		config = c
		api.JobFailure(Job{JobId: 1}, "failed")

		c.Mock.AssertExpectations(t)
		So(postedBody, ShouldEqual, `{"job_id":1,"status":"failure","error_message":"failed"}`)
	})
}

func TestApiJobPostBack(t *testing.T) {
	Convey("Accepting a job post back", t, func() {
		postedBody := ""
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			body, _ := ioutil.ReadAll(r.Body)
			postedBody = string(body)
			w.WriteHeader(200)
		}
		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()
		api = SyswardApi{httpClient: http.Client{}}
		c := new(MockConfig)
		c.On("fetchJobPostbackUrl").Return(server.URL)
		config = c
		api.JobPostBack(Job{JobId: 1})
		c.Mock.AssertExpectations(t)
		So(postedBody, ShouldEqual, `{"job_id":1,"status":"success"}`)
		server.Close()
	})
}

func TestApiCheckIn(t *testing.T) {
	Convey("Geting a succesful a list of jobs", t, func() {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, "[]")
		}
		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()
		api = SyswardApi{httpClient: http.Client{}}
		c := new(MockConfig)
		c.On("fetchJobUrl", getSystemUID()).Return(server.URL)
		config = c

		So(api.GetJobs(), ShouldEqual, "[]")
		c.Mock.AssertExpectations(t)
		server.Close()
	})

	Convey("Getting a list of jobs errors out", t, func() {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			io.WriteString(w, "[]")
		}
		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()
		api = SyswardApi{httpClient: http.Client{}}
		c := new(MockConfig)
		c.On("fetchJobUrl", getSystemUID()).Return(server.URL)
		config = c

		So(api.GetJobs(), ShouldEqual, "")
		c.Mock.AssertExpectations(t)
		server.Close()
	})

	Convey("Getting a list of jobs gives invalid body", t, func() {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
		}
		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()
		api = SyswardApi{httpClient: http.Client{}}
		c := new(MockConfig)
		c.On("fetchJobUrl", getSystemUID()).Return(server.URL)
		config = c

		So(api.GetJobs(), ShouldEqual, "")
		c.Mock.AssertExpectations(t)
		server.Close()

	})

}
