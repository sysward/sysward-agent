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

# Build list of updated packages info
updates_info = []
for pkg in sorted(updates):
    info = dict(
        name=pkg.name,
        candidate_version=pkg.printVer(),
        current_version=versions[pkg.name],
        priority="standard", # can be ignored
        section=pkg.repo.id, # whichever repo is belongs to
        security=False, # did not found any way to retreive the list of security updates.
                        # yum-plugin-security is not supported well in CentOS.
    )
    updates_info.append(info)

# Dump and print list of info
dump = json.dumps(updates_info)

sys.stdout = _stdout
sys.stderr = _stderr
print dump
