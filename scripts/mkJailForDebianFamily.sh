#!/bin/bash
#
# [Description]
# This shell script make rootfs(jail). Rootfs is created as /tmp/mimixbox/jail
# by debootstrap command. It is used as the test environment for chroot
# command and ischroot command.
# This script cannot create rootfs on RHEL-based distributions.
# Only work for Debian-based one.
ROOT_DIR=$(git rev-parse --show-toplevel)
SCRIPT_NAME=$(basename $0)
JAIL="/tmp/mimixbox/jail"
source ${ROOT_DIR}/scripts/libshell.sh

function isDebian() {
    if [ ! -f /etc/debian_version ]; then
        errMsg "This script can only be used with Debian-based distributions."
        exit 1
    fi 
}

function hasDebootstrap() {
    which debootstrap > /dev/null
    local exit_status="$?"
    if [ "${exit_status}" != "0" ]; then
        errMsg "This script use debootstrap command."
        errMsg "[How to install]"
        errMsg " $ sudo apt install debootstrap"
        exit 1
    fi
}

function deleteJailIfNeeded() {
    if [ -e ${JAIL} ]; then
        sudo rm -rf ${JAIL} 
    fi
}

function mkJail() {
    sudo debootstrap bullseye ${JAIL} http://deb.debian.org/debian
}

function cpMimixboxIfNeeded() {
    if [ -f ${ROOT_DIR}/mimixbox ]; then
        cp ${ROOT_DIR}/mimixbox ${JAIL}/usr/bin/.
    fi 
}

isDebian
hasDebootstrap
IS_ROOT=$(isRoot)
if [ "$IS_ROOT" = "1" ]; then
    errMsg "You are not root user."
    exit 1
fi
deleteJailIfNeeded
mkJail
cpMimixboxIfNeeded