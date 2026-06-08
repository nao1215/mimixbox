Setup() {
    export TEST_DIR=/tmp/mimixbox/it
    mkdir -p ${TEST_DIR}
    printf 'x' > ${TEST_DIR}/unlink.txt
}
CleanUp() { rm -rf /tmp/mimixbox/it; }
TestUnlink() {
    unlink /tmp/mimixbox/it/unlink.txt
    test ! -e /tmp/mimixbox/it/unlink.txt && echo gone
}
