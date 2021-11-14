#!/bin/bash -eu
# [Description]
#  This shell script is the installer for the command.
#  This is created to be stored in tar.gz or zip for release files.
ORG_COMMANDS="path serial ghrdc fakemovie mimixbox"
DOC_INSTALL_DIR="/usr/share/doc/mimixbox"

function errMsg() {
    local message="$1"
    echo -n -e "\033[31m\c"
    echo "${message}" >&2
    echo -n -e "\033[m\c"
}

function warnMsg() {
    local message="$1"
    echo -n -e "\033[33m\c"
    echo "${message}"
    echo -n -e "\033[m\c"
}

function isRoot() {
    if [ ${EUID:-${UID}} != 0 ]; then
        errMsg "[Usage]"
        errMsg " $ sudo ./installer.sh"
        exit 1
    fi
}

function installMimixBox() {
	install -v -m 0755 -D mimixbox /usr/local/bin/.
	mimixbox --install /usr/local/bin/.
}

function installManPages() {
    warnMsg "Install man-pages"
    for i in ${ORG_COMMANDS};
    do
        install -v -m 0644 -D docs/man/$i/en/$i.1.gz /usr/share/man/man1/$i.1.gz
        install -v -m 0644 -D docs/man/$i/ja/$i.1.gz /usr/share/man/ja/man1/$i.1.gz
    done
}

function installLicense() {
    warnMsg "Install LICENSE at ${DOC_INSTALL_DIR}"
    which mkdir
    mkdir -p ${DOC_INSTALL_DIR}
    install -v -m 0644 LICENSE ${DOC_INSTALL_DIR}
    install -v -m 0644 NOTICE ${DOC_INSTALL_DIR}
}

isRoot
warnMsg "[Start] Install."
installMimixBox
installManPages
installLicense
warnMsg "[Done]"