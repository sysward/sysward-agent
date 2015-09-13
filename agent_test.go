package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewAgent(t *testing.T) {
	Convey("Setting up a new agent on Ubuntu", t, func() {
		f := new(MockReader)
		f.On("FileExists", "/etc/apt").Return(true)
		r := new(MockRunner)
		r.On("Run", "whoami", []string{}).Return("root", nil)
		agent := NewAgent()
		runner = r
		config_json, _ := ioutil.ReadFile("config.json")
		f.On("ReadFile", "config.json").Return(config_json, nil)
		f.On("FileExists", "/usr/lib/update-notifier/apt-check").Return(true)
		f.On("FileExists", "/etc/apt").Return(true)
		f.On("ReadFile", "config.json").Return(config_json, nil)
		fileReader = f
		agent.Startup()
		So(agent.runner, ShouldHaveSameTypeAs, SyswardRunner{})
		So(agent.fileReader, ShouldHaveSameTypeAs, SyswardFileReader{})
		So(agent.packageManager, ShouldHaveSameTypeAs, DebianPackageManager{})
	})
}

func TestAgentCronInstall(t *testing.T) {
	Convey("Cron job must be installed", t, func() {
		agent := NewAgent()
		f := new(MockReader)
		Convey("Cron is already installed, upstart config doesnt exist", func() {
			f.On("ReadFile", "/etc/crontab").Return([]byte("bin && ./sysward"), nil)
			f.On("FileExists", "/etc/init/sysward-agent.conf").Return(false)
			fileReader = f
			agent.InstallCron()
			f.Mock.AssertExpectations(t)
		})

		// TODO: Interface out the cron installation
		Convey("Cron is not installed, upstart config exists", func() {
			r := new(MockRunner)
			w := new(MockWriter)
			w.On("AppendToFile", "/etc/crontab", "*/5 * * * * root cd /opt/sysward/bin && ./sysward >> /dev/null\n").Return()
			r.On("Run", "/sbin/stop", []string{"sysward-agent"}).Return("", nil)
			r.On("Run", "rm", []string{"-rf", "/etc/init/sysward-agent.conf"}).Return("", nil)
			f.On("ReadFile", "/etc/crontab").Return([]byte(""), nil)
			f.On("FileExists", "/etc/init/sysward-agent.conf").Return(true)
			fileReader = f
			fileWriter = w
			runner = r
			agent.InstallCron()
			f.Mock.AssertExpectations(t)
			r.Mock.AssertExpectations(t)
			w.Mock.AssertExpectations(t)
		})
	})
}

func TestAgentStartup(t *testing.T) {
	Convey("Agent startup should verify root and check pre-req packages", t, func() {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		}
		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()
		r := new(MockRunner)
		f := new(MockReader)
		r.On("Run", "whoami", []string{}).Return("root", nil)
		config_json, _ := ioutil.ReadFile("config.json")
		f.On("FileExists", "/usr/lib/update-notifier/apt-check").Return(true)
		f.On("FileExists", "/etc/apt").Return(true)
		f.On("ReadFile", "config.json").Return(config_json, nil)
		agent := NewAgent()
		runner = r
		fileReader = f
		agent.Startup()
		api = SyswardApi{httpClient: http.Client{}}
		f.Mock.AssertExpectations(t)
		r.Mock.AssertExpectations(t)
	})
}

func TestIfAgentIsRunning(t *testing.T) {
	packageManager = DebianPackageManager{}
	Convey("Checking if the agent is running", t, func() {

		Convey("Agent is running", func() {
			r := new(MockRunner)
			r.On("Run", "ps", []string{"ax"}).Return("./sysward\n./sysward", nil)
			runner = r
			So(CheckIfAgentIsRunning, ShouldPanic)
			r.Mock.AssertExpectations(t)
		})

		Convey("Agent isn't running", func() {
			r := new(MockRunner)
			// We're testing the case where its the only process running
			r.On("Run", "ps", []string{"ax"}).Return("./sysward\n", nil)
			runner = r
			So(CheckIfAgentIsRunning, ShouldNotPanic)
			r.Mock.AssertExpectations(t)
		})

	})

	runner = SyswardRunner{}
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

	packageManager = pm

	config_json, _ := ioutil.ReadFile("config.json")
	f.On("FileExists", "/usr/lib/update-notifier/apt-check").Return(true)
	f.On("ReadFile", "config.json").Return(config_json, nil)
	fileReader = f

	agentData := AgentData{
		Packages:          packageManager.BuildPackageList(),
		SystemUpdates:     packageManager.UpdateCounts(),
		OperatingSystem:   getOsInformation(),
		Sources:           packageManager.GetSourcesList(),
		InstalledPackages: packageManager.BuildInstalledPackageList(),
	}

	a.On("CheckIn", agentData).Return(errors.New("foo"))
	api = a

	Convey("Agent run should checkin, and gather system information", t, func() {
		agent := Agent{}
		agent.Run()
		//r.Mock.AssertExpectations(t)
		//f.Mock.AssertExpectations(t)
		a.Mock.AssertExpectations(t)
		//pm.Mock.AssertExpectations(t)
	})
}
