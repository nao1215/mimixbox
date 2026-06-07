#!/bin/bash -eu
# [Description]
#  This shell script is the installer for the command.
#  This is created to be stored in tar.gz or zip for release files.
DOC_INSTALL_DIR="/usr/share/doc/mimixbox"
ROOT_DIR=$(git rev-parse --show-toplevel)

source ${ROOT_DIR}/scripts/libshell.sh

function installMimixBox() {
	install -v -m 0755 -D mimixbox /usr/local/bin/.
	mimixbox --install /usr/local/bin/.
}

function installLicense() {
    warnMsg "Install LICENSE at ${DOC_INSTALL_DIR}"
    mkdir -p ${DOC_INSTALL_DIR}
    #install -v -m 0644 LICENSE ${DOC_INSTALL_DIR}
    cp -rf licenses ${DOC_INSTALL_DIR}
}

IS_ROOT=$(isRoot)
if [ "$IS_ROOT" = "1" ]; then
    errMsg "[Usage]"
    errMsg " $ sudo ./installer.sh"
    exit 1
fi
warnMsg "[Start] Install."
installMimixBox
installLicense
warnMsg "[Done]"