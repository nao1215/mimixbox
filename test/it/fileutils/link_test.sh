Setup() {
    export TEST_DIR=/tmp/mimixbox/it
    mkdir -p ${TEST_DIR}
    printf 'data' > ${TEST_DIR}/link_src
}
CleanUp() { rm -rf /tmp/mimixbox/it; }
TestLink() {
    link /tmp/mimixbox/it/link_src /tmp/mimixbox/it/link_dst
    cat /tmp/mimixbox/it/link_dst
}
