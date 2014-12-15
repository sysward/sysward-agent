SHELL=/bin/bash
HOSTS = centos6 centos7 ubuntu12 ubuntu14
all: test

test:
	go test -v ./...

build_agent:
	./build.sh

qa: build_agent
	for host in $(HOSTS); do \
		ssh root@$$host.local.sysward.com 'mkdir -p /opt/sysward/bin/'; \
		scp sysward root@$$host.local.sysward.com:/opt/sysward/bin/; \
	done

bump_version:
	~/.rvm/rubies/ruby-2.1.1/bin/ruby -e "f=File.read('version'); File.write('version', f.to_i+1); puts f.to_i+1"

release: build_agent bump_version push

push:
	scp sysward version sysward@web1.sysward.com:~/updates/public
