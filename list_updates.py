#!/bin/env python

try:
    from cStringIO import StringIO
except ImportError:
    from StringIO import StringIO

import sys

_stdout = sys.stdout
_stderr = sys.stderr

sys.stdout = StringIO()
sys.stderr = StringIO()
sys.path.insert(0, '/usr/share/yum-cli')

import yum
import cli
import json
import urllib
import re
import xml.etree.ElementTree as ET
import os.path

# Get verions of all installed packages
base = yum.YumBase()
base.conf.cache = 1
versions = dict()
for pkg in base.rpmdb.returnPackages():
    versions[pkg.name] = pkg.printVer()

# Get list of packages available for update
base_cli = cli.YumBaseCli()
try:
    base_cli.getOptionsConfig(["list", "updates"])
    base_cli.doLock()
    base_cli.doCommands()
    updates = base_cli.doPackageLists(pkgnarrow="updates").updates
finally:
    base_cli.closeRpmDB()
    base_cli.doUnlock()

# Get list of rpms of security updates
ERRATA_URL = "http://cefs.steve-meier.de/errata.latest.xml"

#errataxml = urllib.urlopen(ERRATA_URL).read()
#root = ET.fromstring(errataxml)

#tree = ET.parse("errata.latest.xml")
#root = tree.getroot()

release_file = ""
if os.path.isfile("/etc/os-release"):
    release_file = open("/etc/os-release").read()
if os.path.isfile("/etc/redhat-release"):
    release_file = open("/etc/redhat-release").read()
match = re.search(r"[\w\s]* (\d[\d\.]*)[\w\s]*", release_file)
os_release = match.group(1).split(".")[0]


security_rpms = []

for pkg in sorted(updates):
    d = {}
    (d['n'], d['a'], d['e'], d['v'], d['r']) = pkg.pkgtup
    rpmname = "%(n)s-%(v)s-%(r)s.%(a)s.rpm" % d
    api_url = "http://errata.sysward.com/api/errata/centos/{0}/{1}/{2}".format(os_release, d['n'], rpmname)
    api_response = urllib.urlopen(api_url)
    resp_packages = json.load(api_response)
    if len(resp_packages) > 0:
        security_rpms.append(rpmname)

# Build list of updated packages info
updates_info = []
for pkg in sorted(updates):
    d = {}
    (d['n'], d['a'], d['e'], d['v'], d['r']) = pkg.pkgtup
    rpmname = "%(n)s-%(v)s-%(r)s.%(a)s.rpm" % d
    info = dict(
        name=pkg.name,
        candidate_version=pkg.printVer(),
        current_version=versions[pkg.name],
        priority="standard", # can be ignored
        section=pkg.repo.id, # whichever repo is belongs to
        security=rpmname in security_rpms
    )
    updates_info.append(info)

# Dump and print list of info
dump = json.dumps(updates_info)

sys.stdout = _stdout
sys.stderr = _stderr


print dump
