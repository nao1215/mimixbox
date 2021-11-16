#!/bin/bash
# Support OS and Arch info: https://golang.org/doc/install/source#environment
PROJECT="mimixbox"
ORG_COMMANDS="path serial ghrdc fakemovie"
CWD=$(pwd)
RELEASE=${CWD}/release
BIN_INFO_TXT=${RELEASE}/binary_info.txt
ROOT_DIR=$(git rev-parse --show-toplevel)
MAIN_CODE="${ROOT_DIR}/cmd/mimixbox/main.go"
# Not build windows binary.
OS="linux darwin"
ARCH="386 amd64 arm arm64"
VERSION=$(grep "const version" ${MAIN_CODE} | sed -e "s/const version = \"\(.*\)\"/\1/g")
MANPAGES_DIR="${CWD}/docs/man"

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

function cpManpages() {
    release="$1"

    cd  ${CWD}
    mkdir -p "${release}/docs"
    cp -rf "${MANPAGES_DIR}" "${release}/docs/."

    # Delete unnecessary files
    markdown=$(find ${release} -name "*.md")
    for md in ${markdown};
    do
        rm -f ${md}
    done
}

function cpLicense() {
    release="$1"

    cd  ${CWD}
    cp -f LICENSE ${release}
    cp -f NOTICE ${release}
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
    cpManpages ${release_path}/.

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
    make doc
    mkMacOsAllRelease
    mkLinuxAllRelease
    rmOsDirInRelease
}

main