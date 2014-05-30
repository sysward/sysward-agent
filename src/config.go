package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Config struct {
	Host     string `json:"host"`
	Protocol string `json:"protocol"`
	Interval string `json:"interval"`
}

func NewConfig(filepath string) *Config {
	var err error
	config := new(Config)
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	// config_json := string(file)
	err = json.Unmarshal(file, &config)
	if err != nil {
		panic(err)
	}
	return config
}

func (c Config) fetchJobUrl(uid string) string {
	return fmt.Sprintf("%s://%s/api/v1/jobs?uid=%s", c.Protocol, c.Host, uid)
}

func (c Config) fetchJobPostbackUrl() string {
	return fmt.Sprintf("%s://%s/api/v1/postback", c.Protocol, c.Host)
}

func (c Config) agentCheckinUrl() string {
	return fmt.Sprintf("%s://%s/api/v1/agent", c.Protocol, c.Host)
}
