package main

import (
	"bitbucket.org/sysward/sysward-agent/logging"
	"encoding/json"
	"errors"
	"fmt"
)

type Job struct {
	JobId       int    `json:"job_id"`
	JobType     string `json:"job_type"`
	PackageName string `json:"package_name"`
}

func (job *Job) run() {
	var err error

	if job.JobType == "upgrade-package" {
		logging.LogMsg(fmt.Sprintf("[apt] upgrading: %s", job.PackageName))
		err = packageManager.UpdatePackage(job.PackageName)
	} else if job.JobType == "hold-package" {
		err = packageManager.HoldPackage(job.PackageName)
	} else if job.JobType == "unhold-package" {
		err = packageManager.UnholdPackage(job.PackageName)
	} else {
		err = errors.New(fmt.Sprintf("[job] Unknown job type: %s", job.JobType))
	}

	if err != nil {
		logging.LogMsg(err.Error())
		api.JobFailure(*job, err.Error())

	} else {
		logging.LogMsg(fmt.Sprintf("[job] Posting back for job: %d", job.JobId))
		api.JobPostBack(*job)
	}
}

func runAllJobs(jobs []Job) {
	for index, job := range jobs {
		logging.LogMsg(fmt.Sprintf("Running job %d", index))
		job.run()
		PingApi()
	}
}

func getJobs(config ConfigSettings) []Job {
	var jobs []Job

	jobsResponse := api.GetJobs()

	logging.LogMsg(jobsResponse)
	if jobsResponse == "{}" || jobsResponse == "" {
		return jobs
	} else {
		err := json.Unmarshal([]byte(jobsResponse), &jobs)
		if err != nil {
			panic(err)
		}
	}
	return jobs
}
