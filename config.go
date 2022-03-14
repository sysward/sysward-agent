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
	unregisterAgentUrl() string
	SetProtocol(string)
	Config() ConfigSettings
}

type SyswardConfig struct {
	AgentConfig ConfigSettings
}

func NewConfig(filepath string) ConfigSettings {
	var err error
	config := ConfigSettings{}
	file, err := fileReader.ReadFile(filepath)

	// config_json := string(file)
	err = json.Unmarshal(file, &config)
	if err != nil {
		panic(err)
	}

	return config
}

func (c SyswardConfig) Config() ConfigSettings {
	return c.AgentConfig
}

func (c SyswardConfig) SetProtocol(protocol string) {
	c.AgentConfig.Protocol = protocol
}

func (c SyswardConfig) fetchJobUrl(uid string) string {
	return fmt.Sprintf("%s://%s/api/v1/jobs?uid=%s&api_key=%s", c.AgentConfig.Protocol, c.AgentConfig.Host, uid, c.AgentConfig.ApiKey)
}

func (c SyswardConfig) unregisterAgentUrl() string {
	return fmt.Sprintf("%s://%s/api/v1/unregister?api_key=%s", c.AgentConfig.Protocol, c.AgentConfig.Host, c.AgentConfig.ApiKey)
}

func (c SyswardConfig) fetchJobPostbackUrl() string {
	return fmt.Sprintf("%s://%s/api/v1/postback?api_key=%s", c.AgentConfig.Protocol, c.AgentConfig.Host, c.AgentConfig.ApiKey)
}

func (c SyswardConfig) agentPingUrl() string {
	return fmt.Sprintf("%s://%s/api/v1/ping?api_key=%s", c.AgentConfig.Protocol, c.AgentConfig.Host, c.AgentConfig.ApiKey)
}

func (c SyswardConfig) agentCheckinUrl() string {
	return fmt.Sprintf("%s://%s/api/v1/agent?api_key=%s", c.AgentConfig.Protocol, c.AgentConfig.Host, c.AgentConfig.ApiKey)
}
