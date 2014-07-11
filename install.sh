#!/bin/bash
mkdir -p /opt/sysward/bin
cd /opt/sysward/bin
echo "+ Downloading agent"
wget https://github.com/joshrendek/sysward-agent/releases/download/1.0/sysward
echo "+ Downloading anget config"
wget https://github.com/joshrendek/sysward-agent/releases/download/1.0/config.deploy.json
echo "+ Downloading upstart config"
wget https://github.com/joshrendek/sysward-agent/releases/download/1.0/sysward-agent.deploy.conf
echo "+ Downloading python package reader"
wget https://github.com/joshrendek/sysward-agent/releases/download/1.0/trex.py

chmod +x sysward
mv config.deploy.json config.json
mv sysward-agent.deploy.conf /etc/init/sysward.conf
start sysward
