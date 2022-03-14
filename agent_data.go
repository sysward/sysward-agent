package main

import "encoding/json"

type AgentData struct {
	Packages          []OsPackage     `json:"packages"`
	SystemUpdates     Updates         `json:"system_updates"`
	OperatingSystem   OperatingSystem `json:"operating_system"`
	Sources           []Source        `json:"sources"`
	InstalledPackages []string        `json:"installed_packages"`
	Group             string          `json:"group"`
	CustomHostname    string          `json:"custom_hostname"`
	Hostname          string          `json:"hostname"`
	RebootRequired    bool            `json:"reboot_required"`
}

func (a AgentData) ToJson() (string, error) {
	o, err := json.Marshal(a)
	return string(o), err
}
