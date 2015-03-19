SHELL=/bin/bash
HOSTS = centos6 centos7 ubuntu12 ubuntu14
all: build

test:
	go test -v ./...

build: deps test build_agent

deps:
	go get github.com/tools/godep
	cd src && godep restore

build_agent:
	cd src && GOOS=linux CGO_ENABLED=0 go build -o ../sysward

docker: docker_build docker_run

docker_build:
	docker build --tag="sysward/agent" .

docker_run:
	docker run -v `pwd`:/sysward sysward/agent

qa: build_agent
	for host in $(HOSTS); do \
		ssh root@$$host.local.sysward.com 'mkdir -p /opt/sysward/bin/'; \
		scp sysward root@$$host.local.sysward.com:/opt/sysward/bin/; \
	done

bump_version:
	ruby -e "f=File.read('version'); File.write('version', f.to_i+1); puts f.to_i+1"

release: build_agent bump_version push

deploy:
	echo "Deploying..."

push:
	scp sysward version sysward@web1.sysward.com:~/updates/public
