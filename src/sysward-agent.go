package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log/syslog"
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
	logfile, _ := syslog.New(syslog.LOG_NOTICE, "SYSWARD")
	//logfile := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	pc, _, _, _ := runtime.Caller(1)
	caller := runtime.FuncForPC(pc).Name()
	_, file, line, _ := runtime.Caller(0)
	sp := strings.Split(file, "/")
	shortPath := sp[len(sp)-2 : len(sp)]
	pathLine := fmt.Sprintf("[%s:%d]", shortPath[1], line)
	logString := fmt.Sprintf("%s{%s}:: %s", pathLine, caller, msg)
	logfile.Info(logString)
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
	fileReader = agent.fileReader
	packageManager = agent.packageManager
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
	CheckForUpdate()

	logMsg("package list update - start")
	packageManager.UpdatePackageLists()
	logMsg("package list update - finish")

	logMsg("checking jobs - start")

	jobs := getJobs(config.Config())

	runAllJobs(jobs)

	logMsg("checking jobs - finish")

	counts := packageManager.UpdateCounts()
	operatingSystem := getOsInformation()
	packages := packageManager.BuildPackageList()
	sources := packageManager.GetSourcesList()

	installedPackages := packageManager.BuildInstalledPackageList()

	agentData := AgentData{
		Packages:          packages,
		SystemUpdates:     counts,
		OperatingSystem:   operatingSystem,
		Sources:           sources,
		InstalledPackages: installedPackages,
	}

	err := api.CheckIn(agentData)
	if err != nil {
		logMsg(fmt.Sprintf("[fatal] %s", err))
	}
}

var config Config
var runner Runner
var fileReader SystemFileReader
var packageManager SystemPackageManager
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

	latestVersion, err := strconv.Atoi(string(body))
	if err != nil {
		panic(err)
	}

	if latestVersion > version {
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

func CheckIfAgentIsRunning() {
	procList, _ := runner.Run("ps", "ax")
	out := strings.Split(procList, "\n")
	counter := 0
	for _, proc := range out {
		if strings.Contains(proc, "./sysward") && !strings.Contains(proc, "cd") {
			counter++
		}
	}
	if counter > 1 {
		logMsg(fmt.Sprintf("Sysward already running, exiting. Running: %d", counter))
		os.Exit(1)
	} else {
		logMsg("Sysward is starting.")
	}
}

func main() {
	agent := NewAgent()

	cronString := "*/5 * * * * root cd /opt/sysward/bin && ./sysward\n"
	cronTab, _ := ioutil.ReadFile("/etc/crontab")
	if strings.Contains(string(cronTab), "bin && ./sysward") {
		fmt.Println("+ Cron already installed")
	} else {
		fmt.Println("+ CRON missing - installing")
		f, err := os.OpenFile("/etc/crontab", os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		if _, err = f.WriteString(cronString); err != nil {
			panic(err)
		}
		fmt.Println("CRON installed.")
	}

	if fileReader.FileExists("/etc/init/sysward-agent.conf") {
		fmt.Println("+ Removing upstart config and converting to CRON job...")
		runner.Run("/sbin/stop", "sysward-agent")
		runner.Run("rm", "-rf", "/etc/init/sysward-agent.conf")
		fmt.Println("+ Upstart configs removed and service stopped.")
	}

	CheckIfAgentIsRunning()
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
