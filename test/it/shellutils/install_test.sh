Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/install
    mkdir -p ${TEST_DIR}
    printf 'hello' > ${TEST_DIR}/src
}

CleanUp() {
    rm -rf ${MIMIXBOX_IT_ROOT}/install
}

TestInstallCopyContent() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/install
    install -m 640 ${TEST_DIR}/src ${TEST_DIR}/dst
    printf '%s' "$(< ${TEST_DIR}/dst)"
}

TestInstallMode() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/install
    install -m 640 ${TEST_DIR}/src ${TEST_DIR}/dst2
    stat -c '%a' ${TEST_DIR}/dst2
}

TestInstallDirectory() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/install
    install -d ${TEST_DIR}/a/b/c
    [ -d ${TEST_DIR}/a/b/c ] && echo ok
}

TestInstallNoDest() {
    install only-source
}
