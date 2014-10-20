#!/bin/bash
stop sysward-agent
mkdir -p /opt/sysward/bin
cd /opt/sysward/bin
rm -rf sysward
rm -rf trex.py
echo "+ Downloading agent"
wget -q https://github.com/sysward/sysward-agent/releases/download/1.0/sysward
echo "+ Downloading agent config"
wget -q https://github.com/sysward/sysward-agent/releases/download/1.0/config.deploy.json
echo "+ Downloading python package reader"
wget -q https://github.com/sysward/sysward-agent/releases/download/1.0/trex.py

chmod +x sysward
echo "+ Moving config into place"
mv config.deploy.json config.json
sed -i "s/APIKEY/$API_KEY/g" config.json
echo "+ Running agent one time"
./sysward
echo "+ Logfiles are located at /var/log/sysward.log and tagged with SYSWARD"
