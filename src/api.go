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
	JobFailure(Job, string)
	GetJobs() string
	CheckIn(AgentData) error
}

type SyswardApi struct {
	httpClient *http.Client
}

func (r SyswardApi) CheckIn(agentData AgentData) error {
	client := &r.httpClient
	logMsg("building json - start")
	o, err := agentData.ToJson()
	if err != nil {
		logMsg(fmt.Sprintf("[fatal] %s", err))
		return nil
	}
	logMsg("building json - finish")
	logMsg("posting to api - start")
	post_data := strings.NewReader(o)
	req, err := http.NewRequest("POST", config.agentCheckinUrl(), post_data)
	// formatted_output, _ := json.MarshalIndent(json_output, "", "\t")
	// fmt.Println(string(formatted_output))
	if err != nil {
		logMsg(fmt.Sprintf("[fatal] %s", err))
		return nil
	}
	str, err := client.Do(req)

	if err != nil {
		logMsg(fmt.Sprintf("[fatal] %s", err))
		return nil
	}
	logMsg(string(str.Status))
	logMsg("posting to api - finish")
	return nil
}

func (r SyswardApi) GetJobs() string {
	client := &r.httpClient
	job_url := config.fetchJobUrl(getSystemUID())
	jreq, err := client.Get(job_url)

	if err != nil || jreq.StatusCode != 200 {
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

func (r SyswardApi) JobFailure(job Job, error_string string) {
	logMsg("Posting job FAIL")
	client := &r.httpClient
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
		panic(err)
	}
}

func (r SyswardApi) JobPostBack(job Job) {
	client := &r.httpClient
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
