Setup() { export TEST_DIR=/tmp/mimixbox/it; mkdir -p ${TEST_DIR}; }
CleanUp() { rm -rf /tmp/mimixbox/it; }
TestTruncate() {
    truncate -s 7 /tmp/mimixbox/it/tr_file
    wc -c < /tmp/mimixbox/it/tr_file | tr -d ' '
}
