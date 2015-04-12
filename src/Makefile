SHELL=/bin/bash
HOSTS = centos6 centos7 ubuntu12 ubuntu14
all: build

test:
	go test -v ./...

build: deps test run_agent

deps:
	go get github.com/tools/godep
	godep restore

run_agent:
	go run -race sysward-agent.go agent_data.go api.go apt.go config.go file_reader.go file_writer.go job.go operating_system.go package_manager.go sysward_runner.go types.go yum.go

docker: docker_build docker_run

docker_build:
	docker build --tag="sysward/agent" .

docker_run:
	docker run -i -t -v `pwd`:/sysward sysward/agent
