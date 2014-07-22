package main

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

/* Mocked calls for running system commands */
type MockRunner struct {
	mock.Mock
}

func (r *MockRunner) Run(command string, args ...string) (string, error) {
	// weird way to get around passing in just args to the helper
	pa := []string{}
	pa = append(pa, args...)
	_args := r.Mock.Called(command, pa)
	return _args.String(0), _args.Error(1)
}

/* Mocked calls for reading files and checking if files exist */
type MockReader struct {
	mock.Mock
}

func (r *MockReader) ReadFile(path string) ([]byte, error) {
	args := r.Mock.Called(path)
	return args.Get(0).([]byte), args.Error(1)
}

func (r *MockReader) FileExists(path string) bool {
	args := r.Mock.Called(path)
	return args.Bool(0)
}

/* Mock package manager calls */
type MockPackageManager struct {
	mock.Mock
}

func (r *MockPackageManager) UpdatePackage(name string) error {
	args := r.Mock.Called(name)
	return args.Error(0)
}

func (r *MockPackageManager) HoldPackage(name string) error {
	args := r.Mock.Called(name)
	return args.Error(0)
}

func (r *MockPackageManager) UnholdPackage(name string) error {
	args := r.Mock.Called(name)
	return args.Error(0)
}

func (r *MockPackageManager) BuildPackageList() []OsPackage {
	args := r.Mock.Called()
	return args.Get(0).([]OsPackage)
}

func (r *MockPackageManager) GetSourcesList() []Source {
	args := r.Mock.Called()
	return args.Get(0).([]Source)
}

func (r *MockPackageManager) GetChangelog(name string) string {
	args := r.Mock.Called(name)
	return args.String(0)
}

func (r *MockPackageManager) BuildInstalledPackageList() []string {
	args := r.Mock.Called()
	return args.Get(0).([]string)
}

func (r *MockPackageManager) UpdatePackageLists() error {
	args := r.Mock.Called()
	return args.Error(0)
}

func (r *MockPackageManager) UpdateCounts() Updates {
	args := r.Mock.Called()
	return args.Get(0).(Updates)
}

/* Mock WebAPI calls */
type MockSyswardApi struct {
	mock.Mock
	httpClient *http.Client
}

func (r *MockSyswardApi) JobPostBack(job Job) {
	r.Mock.Called(job)
	return
}

func (r *MockSyswardApi) CheckIn(agentData AgentData) error {
	args := r.Mock.Called(agentData)
	return args.Error(0)
}

func (r *MockSyswardApi) GetJobs() string {
	args := r.Mock.Called()
	return args.String(0)
}

/* Mock Config calls */
type MockConfig struct {
	mock.Mock
	config ConfigSettings
}

func (c *MockConfig) Config() ConfigSettings {
	args := c.Mock.Called()
	return args.Get(0).(ConfigSettings)
}

func (c *MockConfig) agentPingUrl() string {
	args := c.Mock.Called()
	return args.String(0)
}

func (c *MockConfig) fetchJobUrl(uid string) string {
	args := c.Mock.Called(uid)
	return args.String(0)
}

func (c *MockConfig) fetchJobPostbackUrl() string {
	args := c.Mock.Called()
	return args.String(0)
}

func (c *MockConfig) agentCheckinUrl() string {
	args := c.Mock.Called()
	return args.String(0)
}
