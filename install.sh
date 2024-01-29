#!/bin/bash
stop sysward-agent
rm -f /opt/sysward/bin/sysward
rm -f /opt/sysward/bin/trex.py
mkdir -p /opt/sysward/bin
cd /opt/sysward/bin
echo "+ Downloading agent"
wget -q http://updates.sysward.com/sysward
echo "+ Downloading agent config"
wget -q http://updates.sysward.com/config.deploy.json
echo "+ Downloading python package reader"
wget -q http://updates.sysward.com/trex.py
wget -q http://updates.sysward.com/list_updates.py

chmod +x sysward
echo "+ Moving config into place"
mv config.deploy.json config.json
sed -i "s/APIKEY/$API_KEY/g" config.json
echo "+ Running agent one time"
./sysward
echo "+ Logfiles are located at /var/log/syslog and tagged with SYSWARD"
