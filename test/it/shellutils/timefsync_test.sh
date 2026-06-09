Setup() {
    export TEST_DIR=/tmp/mimixbox/it/timefsync
    mkdir -p ${TEST_DIR}
    printf 'data\n' > ${TEST_DIR}/f.txt
}

CleanUp() { rm -rf /tmp/mimixbox/it/timefsync; }

# Use env to invoke the PATH "time" binary regardless of any shell keyword.
TestTimeOutput() {
    env time echo timed 2>/dev/null
}

TestTimeReportsReal() {
    env time echo x 2>&1 1>/dev/null | grep -c real
}

TestFsync() {
    fsync ${TEST_DIR}/f.txt
    echo $?
}

TestFsyncMissing() {
    fsync /no/such/mimixbox/file 2>/dev/null
    echo $?
}
