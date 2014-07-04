package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
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
		api:            SyswardApi{httpClient: &http.Client{}},
	}
	runner = agent.runner
	file_reader = agent.fileReader
	package_manager = agent.packageManager
	api = agent.api
	return &agent
}

func (a *Agent) Startup() {

	verifyRoot()
	checkPreReqs()
	logMsg("pre-reqs verified")
	configSettings := NewConfig("config.json")
	config = SyswardConfig{config: configSettings}
	_, err := time.ParseDuration(config.Config().Interval)
	if err != nil {
		panic(err)
	}
}

func (a *Agent) Run() {
	for {
		logMsg("package list update - start")
		package_manager.UpdatePackageLists()
		logMsg("package list update - finish")

		logMsg("checking jobs - start")

		jobs := getJobs(config.Config())

		runAllJobs(jobs)

		logMsg("checking jobs - finish")

		counts := package_manager.UpdateCounts()
		operating_system := getOsInformation()
		packages := package_manager.BuildPackageList()
		sources := package_manager.GetSourcesList()

		installed_packages := package_manager.BuildInstalledPackageList()

		agentData := AgentData{
			Packages:          packages,
			SystemUpdates:     counts,
			OperatingSystem:   operating_system,
			Sources:           sources,
			InstalledPackages: installed_packages,
		}
		err := api.CheckIn(agentData)
		if err != nil {
			logMsg(fmt.Sprintf("[fatal] %s", err))
			break
		}
		//time.Sleep(interval)
	}
}

var config Config
var runner Runner
var file_reader SystemFileReader
var package_manager SystemPackageManager
var api WebApi

func main() {
	agent := NewAgent()
	agent.Startup()
	agent.Run()
}
