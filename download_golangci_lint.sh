#!/usr/bin/env bash

# The script attempts to download and install the lastest version of the tinygo compiler locally.
# If there is no tinygo already installed then the latest version will be downlaoded and installed.
# If tinygo exists locally but is not at the latest version the latest version will be downloaded nd installed.
# If the latest version of tinygo is already installed locally then noting is changed.

# The script works on a Debian based Linux only.
# It will need updating to support other Linux, BSD's, Windows and MacOS!

# Ask github what the lastest publically released version of timygo is
RELEASE=`curl  "https://api.github.com/repos/golangci/golangci-lint/tags" | jq -r '.[0].name'`
echo The lastest release of golangci-lint is ${RELEASE}
# chop the leading "v" from the version number. We need th everisonin this form for the downlaod URL
RELEASENOV="${RELEASE:1}"
# This is an alternative to the above - without the BASH shell builtin - so it more portable.
#RELEASENOV=`echo ${RELEASE} | cut -c2-`
# is the golangci-lint command instaleld lcoally
GOLANG_CI_LINT_EXE=`command -v golangci-lint`
echo GOLANG_CI_LINT_EXE ${GOLANG_CI_LINT_EXE}
if [[ -z ${GOLANG_CI_LINT_EXE} ]]; # test if string is empty via -z
then
    # no golangci-lint installed at all so download and install the latest version from a deb packae on githuh
    echo No golangci-lint installed locally. Downlaoding and installing golangci-lint ${RELEASE}
    curl -L -o /tmp/golangci-lint.deb https://github.com/golangci/golangci-lint/releases/download/${RELEASE}/golangci-lint-${RELEASENOV}-linux-amd64.deb
    sudo dpkg -i /tmp/golangci-lint.deb # the default Github Actions CI/CD user is a sudo'er. Locally you might need a password.
    rm /tmp/golangci-lint.deb
    exit 0
else
    # some version of golandci-lint is installed
    # a golangvi-lint version string looks like this:
    # golangci-lint has version 1.57.2 built with go1.22.1 from 77a8601a on 2024-03-28T19:01:11Z
    # we want the 4th element (on a zero based array) i.e. 0.31.2
    INSTALLED_GOLANG_CIL_INT_VERSION=`golangci-lint --version | cut -d " " -f 4` 
    echo INSTALLED_GOLANG_CIL_INT_VERSION ${INSTALLED_GOLANG_CIL_INT_VERSION}
    if [[ ${INSTALLED_GOLANG_CIL_INT_VERSION} != ${RELEASENOV} ]];
    then
        curl -L -o /tmp/golangci-lint https://github.com/golangci/golangci-lint/releases/download/${RELEASE}/golangci-lint-${RELEASENOV}-linux-amd64.deb
        sudo dpkg -i /tmp/golangci-lint
        rm /tmp/golangci-lint
    else
        echo Installed golangci-lint is already installed at version ${INSTALLED_GOLANG_CIL_INT_VERSION}
    fi
fi
