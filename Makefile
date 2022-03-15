SHELL=/bin/bash
all: build

test:
	go test -v

build: deps test build_agent cleanup

cleanup:
	rm -rf vendor

deps:
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
	dep ensure

build_agent:
	buildNumber=${GITHUB_RUN_ID: -4}
	echo "*****"
	echo $buildNumber
	echo "*****"
	GOOS=linux CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' \
	  -ldflags "-X main.Version=`date -u +%Y%m%d`$$buildNumber" -o sysward

docker: docker_build docker_run

docker_build:
	docker build --tag="sysward/agent" .

docker_run:
	docker run -e BUILD_NUMBER=$$BUILD_NUMBER -v `pwd`:/go/src/bitbucket.org/sysward/sysward-agent sysward/agent

bump_version:
	ruby -e "f=File.read('version'); File.write('version', f.to_i+1); puts f.to_i+1"

bump_agent_version:
	ruby -e "f=File.read('sysward-agent.go'); v=File.read('version'); f.gsub!('return 38', 'return ' + v); File.write('sysward-agent.go', f);"

release: build_agent bump_version push

deploy:
	echo "Deploying..."
