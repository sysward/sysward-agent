import apt
import apt_pkg
import os
import sys
import gettext
import subprocess
import json

def clean(cache,depcache):
    " unmark (clean) all changes from the given depcache "
    # mvo: looping is too inefficient with the new auto-mark code
    #for pkg in cache.Packages:
    #    depcache.MarkKeep(pkg)
    depcache.init()


def saveDistUpgrade(cache,depcache):
    """ this functions mimics a upgrade but will never remove anything """
    depcache.upgrade(True)
    if depcache.del_count > 0:
        clean(cache,depcache)
        depcache.upgrade()

def isSecurityUpgrade(ver):
    " check if the given version is a security update (or masks one) "
    security_pockets = [("Ubuntu", "%s-security" % DISTRO),
                        ("gNewSense", "%s-security" % DISTRO),
                        ("Debian", "%s-updates" % DISTRO)]
    for (file, index) in ver.file_list:
        for origin, archive in security_pockets:
            if (file.archive == archive and file.origin == origin):
                return True
    return False


output = []

SYNAPTIC_PINFILE = "/var/lib/synaptic/preferences"
DISTRO = subprocess.check_output(
        ["lsb_release", "-c", "-s"],
        universal_newlines=True).strip()

apt_pkg.init()

try:
    cache = apt_pkg.Cache(apt.progress.base.OpProgress())
except SystemError as e:
    sys.stderr.write("E: "+ _("Error: Opening the cache (%s)") % e)
    sys.exit(-1)

depcache = apt_pkg.DepCache(cache)

if os.path.exists(SYNAPTIC_PINFILE):
    depcache.read_pinfile(SYNAPTIC_PINFILE)
    depcache.init()

try:
    saveDistUpgrade(cache,depcache)
except SystemError as e:
    sys.stderr.write("E: "+ _("Error: Marking the upgrade (%s)") % e)
    sys.exit(-1)

# use assignment here since apt.Cache() doesn't provide a __exit__ method
# on Ubuntu 12.04 it looks like
aptcache = apt.Cache()
for pkg in cache.packages:
    if not (depcache.marked_install(pkg) or depcache.marked_upgrade(pkg)):
        continue
    inst_ver = pkg.current_ver
    cand_ver = depcache.get_candidate_ver(pkg)
    if inst_ver == None or cand_ver == None:
        continue
    if cand_ver == inst_ver:
        continue
    record = {
            "name": pkg.name, "security": isSecurityUpgrade(cand_ver), "section": pkg.section,
            "current_version": inst_ver.ver_str, "candidate_version": cand_ver.ver_str,
            "priority": cand_ver.priority_str
            }
    output.append(record)

print json.dumps(output)
