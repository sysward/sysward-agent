package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestBuildingAConfig(t *testing.T) {

	Convey("Given I have a valid configuration", t, func() {

		Convey("The host should be set", nil)

		Convey("The protocol should be https", nil)

		Convey("The interval should be set", nil)

		Convey("The ApiKey should be set", nil)

	})

}

func TestURLBuilding(t *testing.T) {

	Convey("Given I have a valid config", t, func() {

		Convey("fetchJobUrl should be a valid URL", nil)

		Convey("fetchJobPostbackUrl should be a valid URL", nil)

		Convey("agentCheckinUrl should be a valid URL", nil)

	})

}
