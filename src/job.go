package main

import (
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

func getJobs(config ConfigSettings) []Job {
	var jobs []Job

	jobsResponse := api.GetJobs()

	logMsg(jobsResponse)
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
