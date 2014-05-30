package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Job struct {
	JobId       int    `json:"job_id"`
	JobType     string `json:"job_type"`
	PackageName string `json:"package_name"`
}

func (job *Job) run() {
	if job.JobType == "upgrade-package" {
		logMsg(fmt.Sprintf("[apt] upgrading: %s", job.PackageName))
		err := updatePackage(job.PackageName)
		if err != nil {
			panic(err)
		}
	}
	job.postBack()
}

func (job *Job) postBack() {
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

func runAllJobs(jobs []Job) {
	for index, job := range jobs {
		logMsg(fmt.Sprintf("Running job %d", index))
		job.run()
	}
}

func getJobs(config *Config) []Job {
	job_url := config.fetchJobUrl(getSystemUID())

	jreq, err := http.Get(job_url)

	j, err := ioutil.ReadAll(jreq.Body)
	jreq.Body.Close()

	logMsg(string(j))

	var jobs []Job

	err = json.Unmarshal(j, &jobs)
	if err != nil {
		panic(err)
	}

	return jobs
}
