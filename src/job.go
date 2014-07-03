package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Job struct {
	JobId       int    `json:"job_id"`
	JobType     string `json:"job_type"`
	PackageName string `json:"package_name"`
}

func (job *Job) run() {
	var err error

	if job.JobType == "upgrade-package" {
		logMsg(fmt.Sprintf("[apt] upgrading: %s", job.PackageName))
		err = package_manager.UpdatePackage(job.PackageName)
	} else if job.JobType == "hold-package" {
		err = package_manager.HoldPackage(job.PackageName)
	} else if job.JobType == "unhold-package" {
		err = package_manager.UnholdPackage(job.PackageName)
	} else {
		err = errors.New(fmt.Sprintf("[job] Unknown job type: %s", job.JobType))
	}

	if err != nil {
		logMsg(err.Error())
	} else {
		logMsg(fmt.Sprintf("[job] Posting back for job: %d", job.JobId))
		api.JobPostBack(*job)
	}
}

func runAllJobs(jobs []Job) {
	for index, job := range jobs {
		logMsg(fmt.Sprintf("Running job %d", index))
		job.run()
	}
}

func getJobs(config *Config) []Job {
	var jobs []Job
	job_url := config.fetchJobUrl(getSystemUID())

	jreq, err := http.Get(job_url)

	if err != nil {
		logMsg(fmt.Sprintf("Error requesting jobs: %s", err))
		return jobs
	}

	j, err := ioutil.ReadAll(jreq.Body)

	if err != nil {
		logMsg(fmt.Sprintf("Error reading jobs: %s", err))
		return jobs
	}

	jreq.Body.Close()

	logMsg(string(j))

	if string(j) == "{}" {
		return jobs
	} else {
		err = json.Unmarshal(j, &jobs)
		if err != nil {
			panic(err)
		}
	}
	return jobs
}
