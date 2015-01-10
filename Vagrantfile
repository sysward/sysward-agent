# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |con|

  con.vm.define :client_one do |config|
    config.vm.host_name   = "client-1"
    config.vm.box         = 'http://nitron-vagrant.s3-website-us-east-1.amazonaws.com/vagrant_ubuntu_12.04.3_amd64_virtualbox.box'
    config.vm.network     :private_network, ip: "10.10.10.10"
    # config.vm.network     :forwarded_port, guest: 80, host: 8080
    config.vm.network     :forwarded_port, guest: 22, host: 3222, auto: true
  end

  con.vm.define :centos_client do |config|
    config.vm.host_name   = "centos-client"
    config.vm.box         = 'chef/centos-6.5'
    config.vm.network     :private_network, ip: "10.10.10.12"
    # config.vm.network     :forwarded_port, guest: 80, host: 8080
    config.vm.network     :forwarded_port, guest: 22, host: 3224, auto: true
  end

  con.vm.define :centos7_client do |config|
    config.vm.host_name   = "centos-client7"
    config.vm.box         = 'chef/centos-7.0'
    config.vm.network     :private_network, ip: "10.10.10.14", auto_config: false
    # config.vm.network     :forwarded_port, guest: 80, host: 8080
    config.vm.network     :forwarded_port, guest: 22, host: 3227, auto: true
  end

  con.vm.define :client_two do |config|
    config.vm.host_name   = "client-2"
    config.vm.box         = 'ubuntu/trusty64'
    config.vm.network     :private_network, ip: "10.10.10.11"
    # config.vm.network     :forwarded_port, guest: 80, host: 8080
    config.vm.network     :forwarded_port, guest: 22, host: 3223, auto: true
  end

  con.vm.define :client_three do |config|
    config.vm.host_name   = "client-3"
    config.vm.box         = 'ubuntu-server-14.04'
    config.vm.network     :private_network, ip: "10.10.10.09"
    config.vm.network     :forwarded_port, guest: 22, host: 3221, auto: true
  end

end
