SHELL=/bin/bash
HOSTS = 10.10.0.2 10.10.0.3 10.10.0.4 10.10.0.5 10.10.0.13 10.10.0.15 10.10.0.10 10.10.0.9 10.10.0.14 10.10.0.12 10.10.0.11
all: build

test:
	go test -v ./...

build: deps test build_agent

deps:
	go get github.com/tools/godep
	godep restore

build_agent:
	GOOS=linux CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' -o sysward

docker: docker_build docker_run

docker_build:
	docker build --tag="sysward/agent" .

docker_run:
	docker run -v `pwd`:/go/src/bitbucket.org/sysward/sysward-agent sysward/agent

qa:
	for host in $(HOSTS); do \
		scp config.json root@$$host:/opt/sysward/bin/; \
		scp sysward root@$$host:/opt/sysward/bin/; \
	done
qa_run:
	for host in $(HOSTS); do \
		ssh root@$$host "cd /opt/sysward/bin/; ./sysward" ;  \
	done
	wait


bump_version:
	ruby -e "f=File.read('version'); File.write('version', f.to_i+1); puts f.to_i+1"

bump_agent_version:
	ruby -e "f=File.read('sysward-agent.go'); v=File.read('version'); f.gsub!('return 38', 'return ' + v); File.write('sysward-agent.go', f);"

release: build_agent bump_version push

deploy:
	echo "Deploying..."

push:
	scp sysward version sysward@web1.sysward.com:~/updates/public
