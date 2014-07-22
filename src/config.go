package main

import (
	"encoding/json"
	"fmt"
)

type ConfigSettings struct {
	Host     string `json:"host"`
	Protocol string `json:"protocol"`
	Interval string `json:"interval"`
	ApiKey   string `json:"api_key"`
}

type Config interface {
	fetchJobUrl(string) string
	fetchJobPostbackUrl() string
	agentCheckinUrl() string
	agentPingUrl() string
	Config() ConfigSettings
}

type SyswardConfig struct {
	config ConfigSettings
}

func NewConfig(filepath string) ConfigSettings {
	var err error
	config := ConfigSettings{}
	file, err := file_reader.ReadFile(filepath)

	// config_json := string(file)
	err = json.Unmarshal(file, &config)
	if err != nil {
		panic(err)
	}
	return config
}

func (c SyswardConfig) Config() ConfigSettings {
	return c.config
}

func (c SyswardConfig) fetchJobUrl(uid string) string {
	return fmt.Sprintf("%s://%s/api/v1/jobs?uid=%s&api_key=%s", c.config.Protocol, c.config.Host, uid, c.config.ApiKey)
}

func (c SyswardConfig) fetchJobPostbackUrl() string {
	return fmt.Sprintf("%s://%s/api/v1/postback?api_key=%s", c.config.Protocol, c.config.Host, c.config.ApiKey)
}

func (c SyswardConfig) agentPingUrl() string {
	return fmt.Sprintf("%s://%s/api/v1/ping?api_key=%s", c.config.Protocol, c.config.Host, c.config.ApiKey)
}

func (c SyswardConfig) agentCheckinUrl() string {
	return fmt.Sprintf("%s://%s/api/v1/agent?api_key=%s", c.config.Protocol, c.config.Host, c.config.ApiKey)
}
