package main

import (
	"errors"
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewAgent(t *testing.T) {
	Convey("Setting up a new agent", t, func() {
		agent := NewAgent()
		So(agent.runner, ShouldHaveSameTypeAs, SyswardRunner{})
		So(agent.fileReader, ShouldHaveSameTypeAs, SyswardFileReader{})
		So(agent.packageManager, ShouldHaveSameTypeAs, DebianPackageManager{})
	})
}

func TestAgentStartup(t *testing.T) {
	Convey("Agent startup should verify root and check pre-req packages", t, func() {
		r := new(MockRunner)
		f := new(MockReader)
		r.On("Run", "whoami", []string{}).Return("root", nil)
		runner = r
		config_json, _ := ioutil.ReadFile("../config.json")
		f.On("FileExists", "/usr/lib/update-notifier/apt-check").Return(true)
		f.On("ReadFile", "config.json").Return(config_json, nil)
		file_reader = f
		agent := Agent{}
		agent.Startup()
		f.Mock.AssertExpectations(t)
		r.Mock.AssertExpectations(t)
	})
}

func TestAgentRun(t *testing.T) {
	r := new(MockRunner)
	r.On("Run", "whoami", []string{}).Return("root", nil)
	r.On("Run", "lsb_release", []string{"-d"}).Return("Description:    Ubuntu 14.04 LTS", nil)
	r.On("Run", "grep", []string{"MemTotal", "/proc/meminfo"}).Return("MemTotal:        1017764 kB", nil)
	r.On("Run", "grep", []string{"name", "/proc/cpuinfo"}).Return("model name      : Intel(R) Core(TM) i7-4850HQ CPU @ 2.30GHz", nil)
	runner = r

	f := new(MockReader)
	f.On("ReadFile", "/sys/class/dmi/id/product_uuid").Return([]byte("UUID"), nil)

	a := new(MockSyswardApi)
	a.On("GetJobs").Return("")

	pm := new(MockPackageManager)
	pm.On("UpdatePackageLists").Return(nil)
	pm.On("UpdateCounts").Return(Updates{Regular: 0, Security: 0})
	pm.On("BuildPackageList").Return([]OsPackage{})
	pm.On("GetSourcesList").Return([]Source{})
	pm.On("BuildInstalledPackageList").Return([]string{})

	package_manager = pm

	config_json, _ := ioutil.ReadFile("../config.json")
	f.On("FileExists", "/usr/lib/update-notifier/apt-check").Return(true)
	f.On("ReadFile", "config.json").Return(config_json, nil)
	file_reader = f

	agentData := AgentData{
		Packages:          package_manager.BuildPackageList(),
		SystemUpdates:     package_manager.UpdateCounts(),
		OperatingSystem:   getOsInformation(),
		Sources:           package_manager.GetSourcesList(),
		InstalledPackages: package_manager.BuildInstalledPackageList(),
	}

	a.On("CheckIn", agentData).Return(errors.New("foo"))
	api = a

	Convey("Agent run should checkin, and gather system information", t, func() {
		agent := Agent{}
		agent.Run()
	})
}
