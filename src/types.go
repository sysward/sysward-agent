package main

type PatchasaurusOut struct {
  Packages []OsPackage `json:"packages"`
  SystemUpdates Updates `json:"system_updates"`
  OperatingSystem OperatingSystem `json:"operating_system"`
  Sources []Source `json:"sources"`
}

type Interface struct {
  Name string `json:"name"`
  Ips []string `json:"ips"`
  Mac string `json:"mac"`
}

type CPUInformation struct {
  Name string `json:"name"`
  Count int `json:"count"`
}

type MemoryInformation struct {
  Total string `json:"total"`
}

type OperatingSystem struct {
  Name string `json:"name"`
  UID string `json:"uid"`
  Version string `json:"version"`
  Interfaces []Interface `json:"interfaces"`
  Hostname string `json:"hostname"`
  CPUInformation CPUInformation `json:"cpu_information"`
  MemoryInformation MemoryInformation `json:"memory_information"`
}

type OsPackage struct {
  Name string `json:"name"`
  Current_version string `json:"current_version"`
  Candidate_version string `json:"candidate_version"`
  Priority string `json:"priority"`
  Security bool `json:"security"`
  Section string `json:"section"`
}

type Updates struct {
  Regular int `json:"regular"`
  Security int `json:"security"`
}

type Source struct {
  Url string `json:"url"`
  Src bool `json:"src"`
  Channels []string `json:"channels"`
}
