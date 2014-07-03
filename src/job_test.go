package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRunningAJob(t *testing.T) {

	Convey("Given a 'upgrade-package' job", t, func() {

		Convey("The package should be upgraded", func() {
			pm := new(MockPackageManager)
			a := new(MockSyswardApi)
			pm.On("UpdatePackage", "apt").Return(nil)
			job := Job{
				JobId:       1,
				JobType:     "upgrade-package",
				PackageName: "apt",
			}

			a.On("JobPostBack", job).Return()
			package_manager = pm
			api = a
			job.run()
			pm.Mock.AssertExpectations(t)
			a.Mock.AssertExpectations(t)
		})

	})

	Convey("Given a 'hold-package' job", t, func() {

		Convey("The package should be held", nil)

	})

	Convey("Given a 'unhold-package' job", t, func() {

		Convey("The package should be unheld", nil)

	})

}

func TestJobPostback(t *testing.T) {

	Convey("Given the job is successful", t, func() {

		Convey("The job should postback success", nil)

	})

	Convey("Given the job is unccessful", t, func() {

		Convey("The job should not postback", nil)

	})

}

func TestRunningAllJobs(t *testing.T) {

	Convey("Given there are jobs", t, func() {

		Convey("Then all jobs get run", nil)

	})

	Convey("Given there are no jobs", t, func() {

		Convey("Then nothing happens", nil)

	})

}

func TestGettingJobs(t *testing.T) {

	Convey("Given I have a valid configuration", t, nil)

}
