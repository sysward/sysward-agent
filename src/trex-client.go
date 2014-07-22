package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
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

var interval *time.Duration

func (a *Agent) Startup() {
	verifyRoot()
	checkPreReqs()
	logMsg("pre-reqs verified")
	configSettings := NewConfig("config.json")
	config = SyswardConfig{config: configSettings}
}

func PingApi() {
	for {
		client := &http.Client{}
		data := url.Values{}
		data.Set("version", fmt.Sprintf("%d", CurrentVersion()))

		req, err := http.NewRequest("POST", config.agentPingUrl(), bytes.NewBufferString(data.Encode()))
		if err != nil {
			logMsg(fmt.Sprintf("[fatal ping]: %s", err))
		}
		req.Header.Add("X-Sysward-Uid", getSystemUID())
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

		_, err = client.Do(req)
		if err != nil {
			logMsg(fmt.Sprintf("[fatal ping]: %s", err))
		}
		logMsg("[pinging api]")
		time.Sleep(15 * time.Second)
	}
}

func (a *Agent) Run() {
	go PingApi()
	for {
		CheckForUpdate()
		interval, err := time.ParseDuration(config.Config().Interval)

		if err != nil {
			panic(err)
		}

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
		err = api.CheckIn(agentData)
		if err != nil {
			logMsg(fmt.Sprintf("[fatal] %s", err))
			break
		}
		time.Sleep(interval)
	}
}

var config Config
var runner Runner
var file_reader SystemFileReader
var package_manager SystemPackageManager
var api WebApi

func CurrentVersion() int {
	return 23
}

func CheckForUpdate() {
	version := CurrentVersion()
	resp, err := http.Get("http://updates.sysward.com/version")
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	latest_version, err := strconv.Atoi(string(body))
	if err != nil {
		panic(err)
	}

	if latest_version > version {
		logMsg(fmt.Sprintf("Current Version: %d", version))
		logMsg("Downloading latest version: " + string(body))
		runner.Run("mv", "/opt/sysward/bin/sysward", "/opt/sysward/bin/sysward.old")
		runner.Run("curl", "-O", "http://updates.sysward.com/sysward")
		runner.Run("mv", "sysward", "/opt/sysward/bin/")
		runner.Run("chmod", "+x", "/opt/sysward/bin/sysward")
		logMsg("Upgrade finished, exiting")
		os.Exit(0)
	} else {
		logMsg("Versions match - nothing to update")
	}

}

func main() {
	agent := NewAgent()
	agent.Startup()
	agent.Run()
}
