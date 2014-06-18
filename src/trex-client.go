package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

type Runner interface {
	Run(string, ...string) ([]byte, error)
}

func logMsg(msg string) {
	pc, _, _, _ := runtime.Caller(1)
	caller := runtime.FuncForPC(pc).Name()
	_, file, line, _ := runtime.Caller(0)
	sp := strings.Split(file, "/")
	short_path := sp[len(sp)-2 : len(sp)]
	path_line := fmt.Sprintf("[%s/%s:%d]", short_path[0], short_path[1], line)
	log_string := fmt.Sprintf("[%s]%s{%s}:: %s", time.Now(), path_line, caller, msg)
	fmt.Println(log_string)
}

var config *Config
var runner Runner

func main() {
	runner = SyswardRunner{}

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
