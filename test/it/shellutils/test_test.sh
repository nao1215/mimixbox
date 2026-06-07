Setup() {
    export TEST_DIR=/tmp/mimixbox/it/test
    mkdir -p ${TEST_DIR}
    touch ${TEST_DIR}/file.txt
}

CleanUp() {
    rm -rf /tmp/mimixbox/it/test
}

TestTestStringEqual() {
    test abc = abc
    echo "rc=$?"
}

TestTestIntCompare() {
    test 2 -gt 1
    echo "rc=$?"
}

TestTestIntFalse() {
    test 1 -gt 2
    echo "rc=$?"
}

TestTestFileExists() {
    export TEST_DIR=/tmp/mimixbox/it/test
    test -f ${TEST_DIR}/file.txt
    echo "rc=$?"
}

TestTestNegate() {
    test ! -f /no_such_file
    echo "rc=$?"
}
