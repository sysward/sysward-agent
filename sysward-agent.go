package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	//	"./debian"
	"github.com/sysward/sysward-agent/logging"
)

type Agent struct {
	runner         Runner
	fileReader     SystemFileReader
	fileWriter     SystemFileWriter
	packageManager SystemPackageManager
	api            WebApi
	linux          string
}

func NewAgent() *Agent {
	agent = Agent{
		runner:     SyswardRunner{},
		fileReader: SyswardFileReader{},
		fileWriter: SyswardFileWriter{},
		api:        SyswardApi{httpClient: GetHttpClient()},
	}
	runner = agent.runner
	fileReader = agent.fileReader
	fileWriter = agent.fileWriter
	api = agent.api
	return &agent
}

var interval *time.Duration

func (a *Agent) Startup() {
	verifyRoot()

	if fileReader.FileExists("/etc/apt") {
		a.packageManager = DebianPackageManager{}
		a.linux = "debian"
	} else if fileReader.FileExists("/usr/bin/yum") {
		a.packageManager = CentosPackageManager{}
		a.linux = "centos"
	} else if fileReader.FileExists("/usr/bin/zypper") {
		a.packageManager = ZypperPackageManager{}
		a.linux = "suse"
	}

	packageManager = agent.packageManager

	checkPreReqs()
	logging.LogMsg("pre-reqs verified")
	configSettings := NewConfig("config.json")
	config = SyswardConfig{AgentConfig: configSettings}
}

var DefaultDialer = &net.Dialer{Timeout: 2 * time.Second, KeepAlive: 2 * time.Second}

func GetHttpClient() http.Client {
	tr := &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		Dial:                DefaultDialer.Dial,
		TLSHandshakeTimeout: 2 * time.Second,
	}
	if os.Getenv("HTTPS_PROXY") != "" {
		proxyUrl, _ := url.Parse(os.Getenv("HTTPS_PROXY"))
		tr.Proxy = http.ProxyURL(proxyUrl)
	}

	client := http.Client{Transport: tr}
	return client
}

func PingApi() {
	logging.LogMsg(fmt.Sprintf("pinging %s", time.Now()))
	client := GetHttpClient()
	data := url.Values{}
	data.Set("version", fmt.Sprintf("%d", CurrentVersion()))

	req, err := http.NewRequest("POST", config.agentPingUrl(), bytes.NewBufferString(data.Encode()))
	if err != nil {
		logging.LogMsg(fmt.Sprintf("[fatal ping]: %s", err))
		return
	}
	req.Header.Add("X-Sysward-Uid", getSystemUID())
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	_, err = client.Do(req)
	if err != nil {
		logging.LogMsg(fmt.Sprintf("[fatal ping]: %s", err))
	}
	logging.LogMsg(fmt.Sprintf("finished pinging %s", time.Now()))
}

func (a *Agent) Run() {
	var err error

	CheckForUpdate()
	PingApi()

	logging.LogMsg("package list update - start")
	packageManager.UpdatePackageLists()

	logging.LogMsg("package list update - finish")
	PingApi()

	logging.LogMsg("checking jobs - start")

	jobs := getJobs(config.Config())

	runAllJobs(jobs)

	logging.LogMsg("checking jobs - finish")

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

	if len(hostname) > 0 {
		agentData.Hostname = hostname
	}

	if len(group) > 0 {
		agentData.Group = group
	}

	if len(customHostname) > 0 {
		agentData.CustomHostname = customHostname
	}

	err = api.CheckIn(agentData)
	if err != nil {
		logging.LogMsg(fmt.Sprintf("[fatal] %s", err))
	}
	logging.LogMsg("Agent finished")
}

func (a *Agent) InstallCron() {
	cronString := "*/5 * * * * root cd /opt/sysward/bin && ./sysward >> /dev/null\n"
	cronTab, _ := fileReader.ReadFile("/etc/crontab")
	if strings.Contains(string(cronTab), "bin && ./sysward") {
		logging.LogMsg("+ Cron already installed")
	} else {
		logging.LogMsg("+ CRON missing - installing")
		fileWriter.AppendToFile("/etc/crontab", cronString)
		logging.LogMsg("CRON installed.")
	}

	if fileReader.FileExists("/etc/init/sysward-agent.conf") {
		logging.LogMsg("+ Removing upstart config and converting to CRON job...")
		runner.Run("/sbin/stop", "sysward-agent")
		runner.Run("rm", "-rf", "/etc/init/sysward-agent.conf")
		logging.LogMsg("+ Upstart configs removed and service stopped.")
	}
}

var config Config
var runner Runner
var fileReader SystemFileReader
var fileWriter SystemFileWriter
var packageManager SystemPackageManager
var api WebApi
var agent Agent

func CurrentVersion() int {
	return 38
}

func CheckScriptUpdates() {

}

func CheckForUpdate() {
	if os.Getenv("SKIP_UPDATES") == "true" {
		return
	}

	CheckScriptUpdates()

	version := CurrentVersion()
	resp, err := http.Get("https://updates.sysward.com/version")
	if err != nil {
		logging.LogMsg(err.Error())
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
		logging.LogMsg(fmt.Sprintf("Current Version: %d", version))
		logging.LogMsg("Downloading latest version: " + string(body))
		runner.Run("mv", "/opt/sysward/bin/sysward", "/opt/sysward/bin/sysward.old")
		runner.Run("curl", "-O", "http://updates.sysward.com/sysward")
		runner.Run("mv", "sysward", "/opt/sysward/bin/")
		runner.Run("chmod", "+x", "/opt/sysward/bin/sysward")
		logging.LogMsg("Upgrade finished, exiting")
		os.Exit(0)
	} else {
		logging.LogMsg("Versions match - nothing to update")
	}

}

func CheckIfAgentIsRunning() {
	procList, _ := runner.Run("ps", "ax")
	out := strings.Split(procList, "\n")
	counter := 0
	for _, proc := range out {
		if strings.Contains(proc, "./sysward") && !strings.Contains(proc, "cd") && !strings.Contains(proc, "sudo") {
			counter++
		}
	}
	if counter > 1 {
		logging.LogMsg(fmt.Sprintf("Sysward already running, exiting. Running: %d", counter))
		panic("Sysward already running, exiting.")
	} else {
		logging.LogMsg("Sysward is starting.")
	}
}

var group string
var customHostname string
var hostname string

func main() {
	flag.StringVar(&group, "group", "", "join this group automatically or create it")
	flag.StringVar(&customHostname, "custom-hostname", "", "set the custom hostname for this machine")
	flag.StringVar(&hostname, "hostname", "", "set the hostname for this machine")
	flag.Parse()
	agent := NewAgent()

	// TODO: moving this into Startup() caused panics, investigate
	CheckIfAgentIsRunning()
	agent.InstallCron()
	agent.Startup()

	// set Protocol to https if getting a 301 moved
	client := GetHttpClient()
	apiEndpoint := fmt.Sprintf("%s://%s", config.Config().Protocol, config.Config().Host)
	logging.LogMsg("Protocol: " + config.Config().Protocol)
	resp, err := client.Get(apiEndpoint)
	if err != nil {
		logging.LogMsg("Error connecting to the API")
	}

	if err == nil {
		if resp.TLS != nil {
			logging.LogMsg("API using https, switching config protocol")
			newConfig := ConfigSettings{
				Host:     config.Config().Host,
				Protocol: "https",
				Interval: config.Config().Interval,
				ApiKey:   config.Config().ApiKey,
			}
			config = SyswardConfig{AgentConfig: newConfig}
			logging.LogMsg("Config protocol: " + config.Config().Protocol)
		}
	}

	agent.Run()
}
