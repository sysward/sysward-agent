#!/usr/bin/python
import rpm
import os
import os.path
import misc
import i18n
import re
import fnmatch
import stat
import warnings
from subprocess import Popen, PIPE
from rpmUtils import RpmUtilsError
import rpmUtils.miscutils
from rpmUtils.miscutils import flagToString, stringToVersion, compareVerOnly
import Errors
import errno
import struct
from constants import *
from operator import itemgetter

import urlparse
urlparse.uses_fragment.append("media")
from urlgrabber.grabber import URLGrabber, URLGrabError

try:
    import xattr
    if not hasattr(xattr, 'get'):
        xattr = None # This is a "newer" API.
except ImportError:
    xattr = None

# For verify
import pwd
import grp

#  This is for yum-utils/yumdownloader in RHEL-5, where it isn't importing this
# directly but did do "from cli import *", and we did have this in 3.2.22. I
# just _love_ how python re-exports these by default.
from yum.packages import parsePackages
def buildPkgRefDict(pkgs, casematch=True):
    """take a list of pkg objects and return a dict the contains all the possible
       naming conventions for them eg: for (name,i386,0,1,1)
       dict[name] = (name, i386, 0, 1, 1)
       dict[name.i386] = (name, i386, 0, 1, 1)
       dict[name-1-1.i386] = (name, i386, 0, 1, 1)
       dict[name-1] = (name, i386, 0, 1, 1)
       dict[name-1-1] = (name, i386, 0, 1, 1)
       dict[0:name-1-1.i386] = (name, i386, 0, 1, 1)
       dict[name-0:1-1.i386] = (name, i386, 0, 1, 1)
       """
    pkgdict = {}
    for pkg in pkgs:
        (n, a, e, v, r) = pkg.pkgtup
        if not casematch:
            n = n.lower()
            a = a.lower()
            e = e.lower()
            v = v.lower()
            r = r.lower()
        name = n
        nameArch = '%s.%s' % (n, a)
        nameVerRelArch = '%s-%s-%s.%s' % (n, v, r, a)
        nameVer = '%s-%s' % (n, v)
        nameVerRel = '%s-%s-%s' % (n, v, r)
        envra = '%s:%s-%s-%s.%s' % (e, n, v, r, a)
        nevra = '%s-%s:%s-%s.%s' % (n, e, v, r, a)
        for item in [name, nameArch, nameVerRelArch, nameVer, nameVerRel, envra, nevra]:
            if item not in pkgdict:
                pkgdict[item] = []
            pkgdict[item].append(pkg)

    return pkgdict


print buildPkgRefDict()
