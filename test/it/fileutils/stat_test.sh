Setup() { export TEST_DIR=/tmp/mimixbox/it; mkdir -p ${TEST_DIR}; printf 'hello' > ${TEST_DIR}/stat_file; }
CleanUp() { rm -rf /tmp/mimixbox/it; }
TestStatSize() {
    stat -c '%s' /tmp/mimixbox/it/stat_file
}
