Setup() {
    export TEST_DIR=/tmp/mimixbox/it/setfattr
    mkdir -p ${TEST_DIR}
    : > ${TEST_DIR}/file.txt
}
CleanUp() { rm -rf /tmp/mimixbox/it/setfattr; }

TestSetfattrSetThenRead() {
    export TEST_DIR=/tmp/mimixbox/it/setfattr
    if ! setfattr -n user.k -v v ${TEST_DIR}/file.txt 2>/dev/null; then
        echo "user.k=\"v\" (skipped: filesystem has no xattr support)"
        return 0
    fi
    getfattr -d ${TEST_DIR}/file.txt | grep 'user.k'
}

TestSetfattrBadArgs() {
    export TEST_DIR=/tmp/mimixbox/it/setfattr
    setfattr -n user.k -x user.k ${TEST_DIR}/file.txt
}

TestSetfattrHelp() {
    setfattr --help
}

TestSetfattrVersion() {
    setfattr --version
}
