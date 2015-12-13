SHELL=/bin/bash
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
	./qa.sh
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
