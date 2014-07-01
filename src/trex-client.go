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

type Runner interface {
	Run(string, ...string) ([]byte, error)
}

type SystemFileReader interface {
	ReadFile(string) ([]byte, error)
}

type SyswardFileReader struct{}

func (r SyswardFileReader) ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile("../config.json")
}

func logMsg(msg string) {
	logfile := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	logfile.Println(msg)
}

var config *Config
var runner Runner
var file_reader SystemFileReader

func main() {
	runner = SyswardRunner{}
	file_reader = SyswardFileReader{}

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
		updatePackageLists()
		logMsg("package list update - finish")

		client := &http.Client{}

		logMsg("checking jobs - start")

		jobs := getJobs(config)

		runAllJobs(jobs)

		logMsg("checking jobs - finish")

		counts := updateCounts()
		operating_system := getOsInformation()
		logMsg("building package list - start")
		packages := buildPackageList()
		logMsg("building package list - finish")
		logMsg("building sources list - start")
		sources := getSourcesList()
		logMsg("building sources list - finish")

		installed_packages := buildInstalledPackageList()

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
