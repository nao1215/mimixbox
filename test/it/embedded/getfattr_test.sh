Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/getfattr
    mkdir -p ${TEST_DIR}
    : > ${TEST_DIR}/file.txt
}
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}/getfattr; }

# Whether the test filesystem supports user xattrs; "skip" output otherwise.
TestGetfattrRoundTrip() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/getfattr
    if ! setfattr -n user.demo -v hello ${TEST_DIR}/file.txt 2>/dev/null; then
        echo "user.demo (skipped: filesystem has no xattr support)"
        return 0
    fi
    getfattr -d ${TEST_DIR}/file.txt | grep 'user.demo'
}

TestGetfattrNoFile() {
    getfattr
}

TestGetfattrHelp() {
    getfattr --help
}

TestGetfattrVersion() {
    getfattr --version
}
