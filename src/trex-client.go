package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func logMsg(msg string) {
	logfile := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	logfile.Println(msg)
}

type Agent struct {
	runner         Runner
	fileReader     SystemFileReader
	packageManager SystemPackageManager
	api            WebApi
}

func NewAgent() *Agent {
	agent := Agent{
		runner:         SyswardRunner{},
		fileReader:     SyswardFileReader{},
		packageManager: DebianPackageManager{},
		api:            SyswardApi{},
	}
	return &agent
}

func (a *Agent) Startup() {
	verifyRoot()
	checkPreReqs()
	logMsg("pre-reqs verified")
	config = NewConfig("config.json")
	_, err := time.ParseDuration(config.Interval)
	if err != nil {
		panic(err)
	}
}

func (a *Agent) Run() {
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

		//time.Sleep(interval)

	}
}

var config *Config
var runner Runner
var file_reader SystemFileReader
var package_manager SystemPackageManager
var api WebApi

func main() {
	agent := NewAgent()
	runner = agent.runner
	file_reader = agent.fileReader
	package_manager = agent.packageManager
	api = agent.api

	agent.Startup()

	agent.Run()

}
