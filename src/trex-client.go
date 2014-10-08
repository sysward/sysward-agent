package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func logMsg(msg string) {
	logfile := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	pc, _, _, _ := runtime.Caller(1)
	caller := runtime.FuncForPC(pc).Name()
	_, file, line, _ := runtime.Caller(0)
	sp := strings.Split(file, "/")
	short_path := sp[len(sp)-2 : len(sp)]
	path_line := fmt.Sprintf("[%s:%d]", short_path[1], line)
	log_string := fmt.Sprintf("%s{%s}:: %s", path_line, caller, msg)
	logfile.Println(log_string)
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
		api:            SyswardApi{httpClient: GetHttpClient()},
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
	config = SyswardConfig{AgentConfig: configSettings}
}

var timeout = time.Duration(2 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

func GetHttpClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Dial:            dialTimeout,
	}
	client := &http.Client{Transport: tr}
	return client
}

func PingApi() {
	for {
		client := GetHttpClient()
		data := url.Values{}
		data.Set("version", fmt.Sprintf("%d", CurrentVersion()))

		req, err := http.NewRequest("POST", config.agentPingUrl(), bytes.NewBufferString(data.Encode()))
		if err != nil {
			logMsg(fmt.Sprintf("[fatal ping]: %s", err))
			time.Sleep(15 * time.Second)
			continue
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
	return 29
}

func CheckForUpdate() {
	version := CurrentVersion()
	resp, err := http.Get("http://updates.sysward.com/version")
	if err != nil {
		logMsg(err.Error())
		return
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

	// set Protocol to https if getting a 301 moved
	client := GetHttpClient()
	apiEndpoint := fmt.Sprintf("%s://%s", config.Config().Protocol, config.Config().Host)
	logMsg("Protocol: " + config.Config().Protocol)
	resp, err := client.Get(apiEndpoint)
	if err != nil {
		logMsg("Error connecting to the API")
	}

	if err == nil {
		if resp.TLS != nil {
			logMsg("API using https, switching config protocol")
			newConfig := ConfigSettings{
				Host:     config.Config().Host,
				Protocol: "https",
				Interval: config.Config().Interval,
				ApiKey:   config.Config().ApiKey,
			}
			config = SyswardConfig{AgentConfig: newConfig}
			logMsg("Config protocol: " + config.Config().Protocol)
		}
	}

	agent.Run()
}
