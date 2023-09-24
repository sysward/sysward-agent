package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRunningAJob(t *testing.T) {
	pm := new(MockPackageManager)
	a := new(MockSyswardApi)
	job := Job{
		JobId:       1,
		JobType:     "upgrade-package",
		PackageName: "apt",
	}

	Convey("Given a 'upgrade-package' job", t, func() {

		Convey("The package should be upgraded", func() {
			job.JobType = "upgrade-package"
			a.On("JobPostBack", job).Return()
			pm.On("UpdatePackage", "apt").Return(nil)
			packageManager = pm
			api = a
			job.run()
			pm.Mock.AssertExpectations(t)
			a.Mock.AssertExpectations(t)
		})

	})

	Convey("Given a 'hold-package' job", t, func() {

		Convey("The package should be held", func() {
			job.JobType = "hold-package"
			a.On("JobPostBack", job).Return()
			pm.On("HoldPackage", "apt").Return(nil)
			packageManager = pm
			api = a
			job.run()
			pm.Mock.AssertExpectations(t)
			a.Mock.AssertExpectations(t)

		})

	})

	Convey("Given a 'unhold-package' job", t, func() {

		Convey("The package should be unheld", func() {
			job.JobType = "unhold-package"
			a.On("JobPostBack", job).Return()
			pm.On("UnholdPackage", "apt").Return(nil)
			packageManager = pm
			api = a
			job.run()
			pm.Mock.AssertExpectations(t)
			a.Mock.AssertExpectations(t)
		})

	})

	Convey("Given an invalid job type", t, func() {
		job.JobType = "foobar"
		a.AssertNotCalled(t, "JobPostBack", job)
		a.On("JobFailure", job, "[job] Unknown job type: foobar").Return(nil)
		packageManager = pm
		api = a
		job.run()
		pm.Mock.AssertExpectations(t)
		a.Mock.AssertExpectations(t)
	})

}

func TestJobPostback(t *testing.T) {

	Convey("Given the job is successful", t, func() {

		Convey("The job should postback success", func() {

		})

	})

	Convey("Given the job is unccessful", t, func() {

		Convey("The job should not postback", nil)

	})

}

func TestRunningAllJobs(t *testing.T) {
	pm := new(MockPackageManager)
	a := new(MockSyswardApi)
	f := new(MockReader)

	Convey("Given there are jobs", t, func() {

		Convey("Then all jobs get run", func() {
			jobs := []Job{
				Job{
					JobId:       1,
					JobType:     "upgrade-package",
					PackageName: "apt",
				},
				Job{
					JobId:       2,
					JobType:     "upgrade-package",
					PackageName: "foo",
				},
			}
			a.On("JobPostBack", jobs[0]).Return()
			a.On("JobPostBack", jobs[1]).Return()
			f.On("ReadFile", "/opt/sysward/bin/uid").Return([]byte("uid123"), nil)
			f.On("FileExists", "/opt/sysward/bin/uid").Return(true)
			pm.On("UpdatePackage", "apt").Return(nil)
			pm.On("UpdatePackage", "foo").Return(nil)
			packageManager = pm
			api = a
			fileReader = f
			runAllJobs(jobs)
			pm.Mock.AssertExpectations(t)
			a.Mock.AssertExpectations(t)
		})

	})

	Convey("Given there are no jobs", t, func() {

		Convey("Then nothing happens", func() {

		})

	})

}

func TestGettingJobs(t *testing.T) {

	Convey("Given I have a valid configuration", t, func() {
		Convey("There are jobs", func() {
			a := new(MockSyswardApi)
			a.On("GetJobs").Return(`[{"job_id":275,"job_type":"upgrade-package","package_name":"apt"}]`)
			api = a
			jobs := getJobs(config.Config())
			So(jobs[0].JobId, ShouldEqual, 275)
			a.Mock.AssertExpectations(t)
		})

		Convey("There are no jobs", func() {
			a := new(MockSyswardApi)
			a.On("GetJobs").Return(`[]`)
			api = a
			jobs := getJobs(config.Config())
			So(jobs, ShouldBeEmpty)
			a.Mock.AssertExpectations(t)
		})

		Convey("Invalid JSON is sent back", func() {

			a := new(MockSyswardApi)
			a.On("GetJobs").Return(`as0d919{}`)
			api = a
			So(func() { getJobs(config.Config()) }, ShouldPanic)
			a.Mock.AssertExpectations(t)
		})
	})

}
