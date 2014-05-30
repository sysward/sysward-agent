package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func logMsg(msg string) {
	fmt.Println("[", time.Now(), "] - ", msg)
}

var config *Config

func main() {
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

		fmt.Println(jobs)

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

		logMsg("building json - start")
		json_output := PatchasaurusOut{packages, counts, operating_system, sources}
		o, err := json.Marshal(json_output)
		if err != nil {
			panic(err)
		}
		logMsg("building json - finish")

		logMsg("posting to api - start")
		post_data := strings.NewReader(string(o))
		req, err := http.NewRequest("POST", config.agentCheckinUrl(), post_data)
		// formatted_output, _ := json.MarshalIndent(json_output, "", "\t")
		// fmt.Println(string(formatted_output))
		if err != nil {
			panic(err)
		}
		str, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		logMsg(string(str.Status))
		logMsg("posting to api - finish")

		time.Sleep(interval)

	}

}
