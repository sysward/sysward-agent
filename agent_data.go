package main

import "encoding/json"

type AgentData struct {
	Packages          []OsPackage     `json:"packages"`
	SystemUpdates     Updates         `json:"system_updates"`
	OperatingSystem   OperatingSystem `json:"operating_system"`
	Sources           []Source        `json:"sources"`
	InstalledPackages []string        `json:"installed_packages"`
}

func (a AgentData) ToJson() (string, error) {
	o, err := json.Marshal(a)
	return string(o), err
}
