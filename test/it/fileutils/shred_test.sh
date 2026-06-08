Setup() { export TEST_DIR=/tmp/mimixbox/it; mkdir -p ${TEST_DIR}; printf 'secret' > ${TEST_DIR}/shred_file; }
CleanUp() { rm -rf /tmp/mimixbox/it; }
TestShredRemove() {
    shred -u /tmp/mimixbox/it/shred_file
    test ! -e /tmp/mimixbox/it/shred_file && echo gone
}
