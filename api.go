package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"bitbucket.org/sysward/sysward-agent/logging"
)

type WebApi interface {
	JobPostBack(Job)
	JobFailure(Job, string)
	GetJobs() string
	CheckIn(AgentData) error
}

type SyswardApi struct {
	httpClient http.Client
}

func (r SyswardApi) CheckIn(agentData AgentData) error {
	client := r.httpClient
	logging.LogMsg("building json - start")
	o, err := agentData.ToJson()
	if err != nil {
		logging.LogMsg(fmt.Sprintf("[fatal] %s", err))
		return nil
	}
	logging.LogMsg("building json - finish")
	logging.LogMsg("posting to api - start")
	if os.Getenv("DEBUG") == "true" {
		formatted_output, _ := json.MarshalIndent(o, "", "\t")
		fmt.Println(string(formatted_output))
	}
	post_data := strings.NewReader(o)
	req, err := http.NewRequest("POST", config.agentCheckinUrl(), post_data)
	if err != nil {
		logging.LogMsg(fmt.Sprintf("[fatal] %s", err))
		return nil
	}
	str, err := client.Do(req)

	if err != nil {
		logging.LogMsg(fmt.Sprintf("[fatal] %s", err))
		return nil
	}
	logging.LogMsg(string(str.Status))
	logging.LogMsg("posting to api - finish")
	return nil
}

func (r SyswardApi) GetJobs() string {
	client := r.httpClient
	job_url := config.fetchJobUrl(getSystemUID())
	jreq, err := client.Get(job_url)

	if err != nil || jreq.StatusCode != 200 {
		logging.LogMsg(fmt.Sprintf("Error requesting jobs: %s", err))
		return ""
	}

	j, err := ioutil.ReadAll(jreq.Body)

	if err != nil {
		logging.LogMsg(fmt.Sprintf("Error reading jobs: %s", err))
		return ""
	}

	jreq.Body.Close()

	return string(j)
}

func (r SyswardApi) JobFailure(job Job, error_string string) {
	logging.LogMsg("Posting job FAIL")
	client := r.httpClient
	data := struct {
		JobId        int    `json:"job_id"`
		Status       string `json:"status"`
		ErrorMessage string `json:"error_message"`
	}{
		job.JobId,
		"failure",
		error_string,
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
		fmt.Println("Error posting to: " + config.fetchJobPostbackUrl())
		panic(err)
	}
}

func (r SyswardApi) JobPostBack(job Job) {
	client := r.httpClient
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
