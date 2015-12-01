package main

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBuildingAConfig(t *testing.T) {
	fileReader = SyswardFileReader{}
	Convey("Given I have a valid configuration", t, func() {
		configSettings := NewConfig("config.json")
		config = SyswardConfig{AgentConfig: configSettings}
		//Convey("The host should be set", func() {
		//	So(config.Config().Host, ShouldEqual, "192.168.1.57:3000")
		//})

		Convey("The protocol should be http", func() {
			So(config.Config().Protocol, ShouldEqual, "http")
		})

		Convey("The interval should be set", func() {
			So(config.Config().Interval, ShouldEqual, "15s")
		})

		Convey("The ApiKey should be set", func() {
			So(config.Config().ApiKey, ShouldNotBeNil)
		})

	})

}

func TestURLBuilding(t *testing.T) {
	uid := "abc"
	fileReader = SyswardFileReader{}
	Convey("Given I have a valid config", t, func() {
		configSettings := NewConfig("config.json")
		config = SyswardConfig{AgentConfig: configSettings}

		Convey("fetchJobUrl should be a valid URL", func() {
			_url := config.fetchJobUrl(uid)
			u, _ := url.Parse(_url)
			So(u.Path, ShouldEqual, "/api/v1/jobs")
			So(u.RawQuery, ShouldContainSubstring, "uid=abc&api_key=")
		})

		Convey("fetchJobPostbackUrl should be a valid URL", func() {
			_url := config.fetchJobPostbackUrl()
			u, _ := url.Parse(_url)
			So(u.Path, ShouldEqual, "/api/v1/postback")
			So(u.RawQuery, ShouldContainSubstring, "api_key=")
		})

		Convey("agentCheckinUrl should be a valid URL", func() {
			_url := config.agentCheckinUrl()
			u, _ := url.Parse(_url)
			So(u.Path, ShouldEqual, "/api/v1/agent")
			So(u.RawQuery, ShouldContainSubstring, "api_key=")
		})

	})

}
