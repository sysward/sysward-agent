#!/bin/bash
mkdir -p /opt/sysward/bin
cd /opt/sysward/bin
echo "+ Downloading agent"
wget -q https://github.com/joshrendek/sysward-agent/releases/download/1.0/sysward
echo "+ Downloading anget config"
wget -q https://github.com/joshrendek/sysward-agent/releases/download/1.0/config.deploy.json
echo "+ Downloading upstart config"
wget -q https://github.com/joshrendek/sysward-agent/releases/download/1.0/sysward-agent.deploy.conf
echo "+ Downloading python package reader"
wget -q https://github.com/joshrendek/sysward-agent/releases/download/1.0/trex.py

chmod +x sysward
echo "+ Moving config into place"
mv config.deploy.json config.json
sed -i "s/APIKEY/$API_KEY/g" config.json
echo "+ Installing upstart config"
mv sysward-agent.deploy.conf /etc/init/sysward-agent.conf
echo "+ Running 'start sysward-agent' to start agent"
start sysward-agent
echo "+ Logfiles are located at /var/log/upstart/sysward-agent.log"
