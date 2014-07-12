all: test

test: 
	go test -v ./...

build_agent: 
	./build.sh


bump_version:
	~/.rvm/rubies/ruby-2.1.1/bin/ruby -e "f=File.read('version'); File.write('version', f.to_i+1); puts f.to_i+1"

release: build_agent bump_version push


push:
	scp version sysward sysward@web1.sysward.com:~/updates/public

