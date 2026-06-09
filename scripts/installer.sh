#!/bin/bash -eu
# [Description]
#  Self-contained installer for MimixBox, shipped inside the release archive.
#  It resolves everything relative to its own location (with a current-directory
#  fallback for `make install` from a Git checkout), so it works from a plain
#  extracted archive that has no Git metadata.
DOC_INSTALL_DIR="/usr/share/doc/mimixbox"
INSTALL_DIR="/usr/local/bin"
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)

source "${SCRIPT_DIR}/libshell.sh"

# resolveAsset prints the path to NAME, preferring the script's own directory
# (release archive layout) and falling back to the current directory (running
# from a Git checkout via `make install`). It prints nothing when not found.
function resolveAsset() {
    local name="$1"
    if [ -e "${SCRIPT_DIR}/${name}" ]; then
        echo "${SCRIPT_DIR}/${name}"
    elif [ -e "./${name}" ]; then
        echo "./${name}"
    fi
}

function installMimixBox() {
    local bin
    bin=$(resolveAsset mimixbox)
    if [ -z "${bin}" ]; then
        errMsg "mimixbox binary not found next to installer.sh or in the current directory"
        exit 1
    fi
    install -v -m 0755 -D "${bin}" "${INSTALL_DIR}/mimixbox"
    # Invoke the exact binary just installed (not a PATH lookup) so the applet
    # symlinks are guaranteed to target ${INSTALL_DIR}/mimixbox.
    "${INSTALL_DIR}/mimixbox" --install "${INSTALL_DIR}/."
}

function installLicense() {
    local license licenses_dir
    license=$(resolveAsset LICENSE)
    licenses_dir=$(resolveAsset licenses)
    mkdir -p "${DOC_INSTALL_DIR}"
    if [ -n "${license}" ]; then
        warnMsg "Install LICENSE at ${DOC_INSTALL_DIR}"
        cp -f "${license}" "${DOC_INSTALL_DIR}/."
    fi
    if [ -n "${licenses_dir}" ]; then
        warnMsg "Install dependency licenses at ${DOC_INSTALL_DIR}"
        cp -rf "${licenses_dir}" "${DOC_INSTALL_DIR}/."
    fi
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
