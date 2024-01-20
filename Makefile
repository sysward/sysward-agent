SHELL=/bin/bash
BUILD_NUMBER=${GITHUB_RUN_ID}
all: build

test:
	go test -v ./...

build: test build_agent

build_agent:
	@echo "*****"
	@echo ${BUILD_NUMBER}
	@echo "*****"
	GOOS=linux CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' \
	  -ldflags "-X main.Version=`date -u +%Y%m%d`${BUILD_NUMBER}" -o sysward
	GOOS=linux CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' \
	  -ldflags "-X main.Version=`date -u +%Y%m%d`${BUILD_NUMBER}" -o sysward_x86_64
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' \
	  -ldflags "-X main.Version=`date -u +%Y%m%d`${BUILD_NUMBER}" -o sysward_arm64
	GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' \
	  -ldflags "-X main.Version=`date -u +%Y%m%d`${BUILD_NUMBER}" -o sysward_armv7l
	GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' \
	  -ldflags "-X main.Version=`date -u +%Y%m%d`${BUILD_NUMBER}" -o sysward_armv6l
	echo -n `date -u +%Y%m%d`${BUILD_NUMBER} > version

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
