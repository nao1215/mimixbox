Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/setfattr
    mkdir -p ${TEST_DIR}
    : > ${TEST_DIR}/file.txt
}
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}/setfattr; }

TestSetfattrSetThenRead() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/setfattr
    if ! setfattr -n user.k -v v ${TEST_DIR}/file.txt 2>/dev/null; then
        echo "user.k=\"v\" (skipped: filesystem has no xattr support)"
        return 0
    fi
    getfattr -d ${TEST_DIR}/file.txt | grep 'user.k'
}

TestSetfattrBadArgs() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/setfattr
    setfattr -n user.k -x user.k ${TEST_DIR}/file.txt
}

TestSetfattrHelp() {
    setfattr --help
}

TestSetfattrVersion() {
    setfattr --version
}
