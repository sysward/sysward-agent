package main

type SystemPackageManager interface {
	UpdatePackage(string) error
	HoldPackage(string) error
	UnholdPackage(string) error
	BuildPackageList() []OsPackage
	GetSourcesList() []Source
	GetChangelog(string) string
	BuildInstalledPackageList() []string
	UpdatePackageLists() error
	UpdateCounts() Updates
}
