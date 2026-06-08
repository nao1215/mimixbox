Setup() {
    export TEST_DIR=/tmp/mimixbox/it
    mkdir -p ${TEST_DIR}
    printf 'apple\nbanana\n' > ${TEST_DIR}/comm_a.txt
    printf 'banana\ncherry\n' > ${TEST_DIR}/comm_b.txt
}
CleanUp() {
    rm -rf /tmp/mimixbox/it
}
TestCommBoth() {
    comm -1 -2 /tmp/mimixbox/it/comm_a.txt /tmp/mimixbox/it/comm_b.txt
}
