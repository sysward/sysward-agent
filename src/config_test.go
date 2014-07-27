package main

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBuildingAConfig(t *testing.T) {
	file_reader = SyswardFileReader{}
	Convey("Given I have a valid configuration", t, func() {
		configSettings := NewConfig("../config.json")
		config = SyswardConfig{AgentConfig: configSettings}
		Convey("The host should be set", func() {
			So(config.Config().Host, ShouldEqual, "10.0.2.2:5000")
		})

		Convey("The protocol should be http", func() {
			So(config.Config().Protocol, ShouldEqual, "http")
		})

		Convey("The interval should be set", func() {
			So(config.Config().Interval, ShouldEqual, "15s")
		})

		Convey("The ApiKey should be set", func() {
			So(config.Config().ApiKey, ShouldEqual, "d4b6c0ebf64456b1bec50cc679b146ed77b88195d681b96a902d15299c1ed51a")
		})

	})

}

func TestURLBuilding(t *testing.T) {
	uid := "abc"
	file_reader = SyswardFileReader{}
	Convey("Given I have a valid config", t, func() {
		configSettings := NewConfig("../config.json")
		config = SyswardConfig{AgentConfig: configSettings}

		Convey("fetchJobUrl should be a valid URL", func() {
			_url := config.fetchJobUrl(uid)
			u, _ := url.Parse(_url)
			So(u.Path, ShouldEqual, "/api/v1/jobs")
			So(u.RawQuery, ShouldEqual, "uid=abc&api_key=d4b6c0ebf64456b1bec50cc679b146ed77b88195d681b96a902d15299c1ed51a")
		})

		Convey("fetchJobPostbackUrl should be a valid URL", func() {
			_url := config.fetchJobPostbackUrl()
			u, _ := url.Parse(_url)
			So(u.Path, ShouldEqual, "/api/v1/postback")
			So(u.RawQuery, ShouldEqual, "api_key=d4b6c0ebf64456b1bec50cc679b146ed77b88195d681b96a902d15299c1ed51a")
		})

		Convey("agentCheckinUrl should be a valid URL", func() {
			_url := config.agentCheckinUrl()
			u, _ := url.Parse(_url)
			So(u.Path, ShouldEqual, "/api/v1/agent")
			So(u.RawQuery, ShouldEqual, "api_key=d4b6c0ebf64456b1bec50cc679b146ed77b88195d681b96a902d15299c1ed51a")
		})

	})

}
