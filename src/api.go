package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type WebApi interface {
	JobPostBack(Job)
	GetJobs() string
}

type SyswardApi struct{}

func (r SyswardApi) GetJobs() string {
	job_url := config.fetchJobUrl(getSystemUID())

	jreq, err := http.Get(job_url)

	if err != nil {
		logMsg(fmt.Sprintf("Error requesting jobs: %s", err))
		return ""
	}

	j, err := ioutil.ReadAll(jreq.Body)

	if err != nil {
		logMsg(fmt.Sprintf("Error reading jobs: %s", err))
		return ""
	}

	jreq.Body.Close()

	return string(j)
}

func (r SyswardApi) JobPostBack(job Job) {
	client := &http.Client{}
	data := struct {
		JobId  int    `json:"job_id"`
		Status string `json:"status"`
	}{
		job.JobId,
		"success",
	}
	o, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	post_data := strings.NewReader(string(o))
	req, err := http.NewRequest("POST", config.fetchJobPostbackUrl(), post_data)
	req.Header.Add("X-Sysward-Uid", getSystemUID())
	_, err = client.Do(req)
	if err != nil {
		panic(err)
	}
}
