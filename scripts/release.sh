#!/bin/bash
# Support OS and Arch info: https://golang.org/doc/install/source#environment
PROJECT="mimixbox"
CWD=$(pwd)
RELEASE=${CWD}/release
BIN_INFO_TXT=${RELEASE}/binary_info.txt
ROOT_DIR=$(git rev-parse --show-toplevel)
# Not build windows binary.
OS="linux darwin"
ARCH="386 amd64 arm arm64"
# Canonical version source: the latest git tag (without its leading "v"). This
# matches the value `make build` and GoReleaser inject into the binary, instead
# of the long-removed `const version` that this script used to scrape.
VERSION=$(git -C "${ROOT_DIR}" describe --tags --abbrev=0 2>/dev/null | sed 's/^v//')
VERSION=${VERSION:-dev}

function mkReleaseDir() {
    cd  ${CWD}
    for os in $OS;
    do
        mkdir -p ${RELEASE}/$os
        for arch in $ARCH;
        do
            mkdir -p ${RELEASE}/$os/${PROJECT}-${VERSION}-${os}-${arch}
        done
    done
}

function cpLicense() {
    release="$1"

    cd  ${CWD}
    cp -f LICENSE ${release}
    cp -rf licenses ${release}
}

function cpInstaller() {
    release="$1"

    cd ${CWD}
    cp -f scripts/installer.sh ${release}
}

# TODO: Functionize
function mkRelease() {
    cd ${CWD}
    os="$1"
    arch="$2"
    release_dir=${PROJECT}-${VERSION}-${os}-${arch}
    release_path="${RELEASE}/$os/${release_dir}"
    tarball="${release_dir}.tar.gz"
    zip_file="${release_dir}.zip"

    echo "---Make release files for OS=$os Architecture=$arch" >> ${BIN_INFO_TXT}
    GOOS=$os GOARCH=$arch make build
    echo -n "   " >> ${BIN_INFO_TXT}
    file ${CWD}/${PROJECT} >> ${BIN_INFO_TXT}
    echo "" >> ${BIN_INFO_TXT}
    mv ${CWD}/${PROJECT} ${release_path}/.
    cpLicense ${release_path}/.
    cpInstaller ${release_path}/.

    cd ${RELEASE}/$os/
    tar cvfz "${tarball}" "${release_dir}"
    zip "${zip_file}" -r "${release_dir}"

    mv "${tarball}" "${RELEASE}/."
    mv "${zip_file}" "${RELEASE}/."
    rm -rf ${release_dir}
    cd ${CWD}
}

function mkLinuxRelease() {
    cd ${CWD}
    os="linux"
    arch="$1"

    mkRelease ${os} ${arch}
}

function mkMacOsRelease() {
    cd ${CWD}
    os="darwin"
    arch="$1"

    mkRelease ${os} ${arch}
}

function mkLinuxAllRelease() {
    arch="386 amd64 arm arm64"
    for i in ${arch};
    do
        mkLinuxRelease "$i"
    done
}

function mkMacOsAllRelease() {
    # Not support 386, arm, arm64.
    arch="amd64"
    for i in ${arch};
    do
        mkMacOsRelease "$i"
    done
}

function mkSrcRelease() {
    TMP=$(mktemp -d)
    code="${PROJECT}-${VERSION}-src"
    #tarball="$code.tar.gz"
    zip_file="$code.zip"

    cd  ${CWD}
    cp -r ${CWD}/../${PROJECT} ${TMP}/.
    mkdir -p ${RELEASE}
    mv ${TMP}/${PROJECT} ${RELEASE}/.

    cd  ${RELEASE}
    #tar cvfz ${tarball} "${PROJECT}"
    zip ${zip_file} -r "${PROJECT}"
    rm -rf "${RELEASE}/${PROJECT}"
    cd  ${CWD}
}

function rmOsDirInRelease() {
    for os in ${OS};
    do
        rm -rf ${RELEASE}/$os
    done
}

function main() {
    cd ${CWD}
    make clean
    #mkSrcRelease

    touch ${BIN_INFO_TXT}
    mkReleaseDir
    mkMacOsAllRelease
    mkLinuxAllRelease
    rmOsDirInRelease
}

main