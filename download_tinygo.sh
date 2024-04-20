#!/usr/bin/env bash

# The script attempts to download and install the lastest version of the tinygo compiler locally.
# If there is no tinygo already installed then the latest version will be downlaoded and installed.
# If tinygo exists locally but is not at the latest version the latest version will be downloaded nd installed.
# If the latest version of tinygo is already installed locally then noting is changed.

# The script works on a Debian based Linux only.
# It will need updating to support other Linux, BSD's, Windows and MacOS!

# Ask github what the lastest publically released version of timygo is
RELEASE=`curl  "https://api.github.com/repos/tinygo-org/tinygo/tags" | jq -r '.[0].name'`
echo The lastest release of TinyGo is ${RELEASE}
# chop the leading "v" from the version number. We need th everisonin this form for the downlaod URL
RELEASENOV="${RELEASE:1}"
# This is an alternative to the above - without the BASH shell builtin - so it more portable.
#RELEASENOV=`echo ${RELEASE} | cut -c2-`
# is there a tinygo nstalled locally?
TINY_GO_EXE=`command -v tinygo`
if [[ -z ${TINY_GO_EXE} ]]; # is string empty via -z
then
    # no tinygo installed at all so download and install the latest version from a deb packae on github
    echo No tinygo installed locally. Downlaoding and installing tinygo ${RELEASE}
    curl -L -o /tmp/tinygo.deb https://github.com/tinygo-org/tinygo/releases/download/${RELEASE}/tinygo_${RELEASENOV}_amd64.deb
    sudo dpkg -i /tmp/tinygo.deb # the default Github Actions CI/CD user is a sudo'er. Locally you might need a password.
    rm /tmp/tinygo.deb
    exit 0
else
    # some version of tinygo is installed
    # a tinygo version string looks like this:
    # tinygo version 0.31.2 linux/amd64 (using go version go1.22.1 and LLVM version 17.0.1)
    # we want the 3rd element (on a zero based array) i.e. 0.31.2
    INSTALLED_TINY_GO_VERSION=`tinygo version | cut -d " " -f 3` 
    echo Found tinygo version ${INSTALLED_TINY_GO_VERSION}
    # do we have the latest version of tinygo installed locally?
    if [[ ${INSTALLED_TINY_GO_VERSION} != ${RELEASENOV} ]];
    then
        echo Upgrading tinygo to version${RELEASE}
        curl -L -o /tmp/tinygo.deb https://github.com/tinygo-org/tinygo/releases/download/${RELEASE}/tinygo_${RELEASENOV}_amd64.deb
        sudo dpkg -i /tmp/tinygo.deb
        rm /tmp/tinygo.deb
    else
        echo Installed tinygo is already installed at lastest version ${INSTALLED_TINY_GO_VERSION}
    fi
fi
