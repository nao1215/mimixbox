Setup() {
    export TEST_DIR=/tmp/mimixbox/it/fallocate
    mkdir -p ${TEST_DIR}
}

CleanUp() { rm -rf /tmp/mimixbox/it/fallocate; }

TestSetsid() {
    setsid echo "session ok"
}

TestFallocate() {
    fallocate -l 4096 ${TEST_DIR}/f
    wc -c < ${TEST_DIR}/f
}

TestFallocateNoLength() {
    fallocate ${TEST_DIR}/x 2>/dev/null
    echo $?
}
