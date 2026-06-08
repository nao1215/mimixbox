Setup() {
    export TEST_DIR=/tmp/mimixbox/it/install
    mkdir -p ${TEST_DIR}
    printf 'hello' > ${TEST_DIR}/src
}

CleanUp() {
    rm -rf /tmp/mimixbox/it/install
}

TestInstallCopyContent() {
    export TEST_DIR=/tmp/mimixbox/it/install
    install -m 640 ${TEST_DIR}/src ${TEST_DIR}/dst
    printf '%s' "$(< ${TEST_DIR}/dst)"
}

TestInstallMode() {
    export TEST_DIR=/tmp/mimixbox/it/install
    install -m 640 ${TEST_DIR}/src ${TEST_DIR}/dst2
    stat -c '%a' ${TEST_DIR}/dst2
}

TestInstallDirectory() {
    export TEST_DIR=/tmp/mimixbox/it/install
    install -d ${TEST_DIR}/a/b/c
    [ -d ${TEST_DIR}/a/b/c ] && echo ok
}

TestInstallNoDest() {
    install only-source
}
