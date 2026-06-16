# shellcheck shell=sh
# Helper for spec/ls_gnu_spec.sh: GNU ls presentation flags (#722-#726).

GnuSetup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/ls_gnu
    rm -rf "${TEST_DIR}"
    mkdir -p "${TEST_DIR}/adir"
    # small.txt: 10 bytes, big.txt: 5000 bytes (distinct sizes).
    printf 'xxxxxxxxxx' > "${TEST_DIR}/small.txt"
    dd if=/dev/zero of="${TEST_DIR}/big.txt" bs=1 count=5000 2>/dev/null
    : > "${TEST_DIR}/a.log"
    : > "${TEST_DIR}/tmpfile"
    printf '#!/bin/sh\n' > "${TEST_DIR}/run.sh"
    chmod 0755 "${TEST_DIR}/run.sh"
    ln -s small.txt "${TEST_DIR}/link"
}

GnuCleanUp() { rm -rf "${MIMIXBOX_IT_ROOT}/ls_gnu"; }

TestColorAlways() { ls --color=always "${TEST_DIR}"; }
TestColorNever() { ls --color=never "${TEST_DIR}"; }
TestClassify() { ls -F "${TEST_DIR}"; }
TestFileType() { ls --file-type "${TEST_DIR}"; }
TestIndicatorSlash() { ls --indicator-style=slash "${TEST_DIR}"; }
TestSortSize() { ls --sort=size "${TEST_DIR}"; }
TestGroupDirs() { ls --group-directories-first "${TEST_DIR}"; }
TestIgnoreLog() { ls --ignore='*.log' "${TEST_DIR}"; }
TestHideTmp() { ls --hide='tmp*' "${TEST_DIR}"; }
TestHideTmpWithAll() { ls -a --hide='tmp*' "${TEST_DIR}"; }
TestInode() { ls -i "${TEST_DIR}/small.txt"; }
TestBlockSize() { ls -l -k "${TEST_DIR}/big.txt"; }
