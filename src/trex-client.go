package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type SystemFileReader interface {
	ReadFile(string) ([]byte, error)
	FileExists(string) bool
}

type WebApi interface {
	JobPostBack(Job)
	GetJobs() string
}

type SystemPackageManager interface {
	UpdatePackage(string) error
	HoldPackage(string) error
	UnholdPackage(string) error
	BuildPackageList() []OsPackage
	GetSourcesList() []Source
	GetChangelog(string) string
	BuildInstalledPackageList() []string
	UpdatePackageLists() error
	UpdateCounts() Updates
}

type SyswardFileReader struct{}

type SyswardApi struct{}

func (r SyswardFileReader) ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func (r SyswardFileReader) FileExists(path string) bool {
	if _, err := os.Stat("/usr/lib/update-notifier/apt-check"); os.IsNotExist(err) {
		return false
	}
	return true
}

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

func logMsg(msg string) {
	logfile := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	logfile.Println(msg)
}

var config *Config
var runner Runner
var file_reader SystemFileReader
var package_manager SystemPackageManager
var api WebApi

func main() {
	runner = SyswardRunner{}
	file_reader = SyswardFileReader{}
	package_manager = DebianPackageManager{}
	api = SyswardApi{}

	out, err := runner.Run("echo", "hello")
	fmt.Println(string(out))

	verifyRoot()
	logMsg("root verified")

	checkPreReqs()
	logMsg("pre-reqs verified")

	config = NewConfig("config.json")
	interval, err := time.ParseDuration(config.Interval)
	if err != nil {
		panic(err)
	}

	for {

		logMsg("package list update - start")
		package_manager.UpdatePackageLists()
		logMsg("package list update - finish")

		client := &http.Client{}

		logMsg("checking jobs - start")

		jobs := getJobs(config)

		runAllJobs(jobs)

		logMsg("checking jobs - finish")

		counts := package_manager.UpdateCounts()
		operating_system := getOsInformation()
		logMsg("building package list - start")
		packages := package_manager.BuildPackageList()
		logMsg("building package list - finish")
		logMsg("building sources list - start")
		sources := package_manager.GetSourcesList()
		logMsg("building sources list - finish")

		installed_packages := package_manager.BuildInstalledPackageList()

		logMsg("building json - start")
		json_output := PatchasaurusOut{packages, counts, operating_system, sources, installed_packages}
		o, err := json.Marshal(json_output)
		if err != nil {
			logMsg(fmt.Sprintf("[fatal] %s", err))
			continue
		}
		logMsg("building json - finish")

		logMsg("posting to api - start")
		post_data := strings.NewReader(string(o))
		req, err := http.NewRequest("POST", config.agentCheckinUrl(), post_data)
		// formatted_output, _ := json.MarshalIndent(json_output, "", "\t")
		// fmt.Println(string(formatted_output))
		if err != nil {
			logMsg(fmt.Sprintf("[fatal] %s", err))
			continue
		}
		str, err := client.Do(req)
		if err != nil {
			logMsg(fmt.Sprintf("[fatal] %s", err))
			continue
		}
		logMsg(string(str.Status))
		logMsg("posting to api - finish")

		time.Sleep(interval)

	}

}
