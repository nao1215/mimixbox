Setup() {
    export TEST_DIR=/tmp/mimixbox/it/cpio
    mkdir -p ${TEST_DIR}/in
    printf 'payload' > ${TEST_DIR}/in/file.txt
}
CleanUp() { rm -rf /tmp/mimixbox/it/cpio; }

TestCpioRoundTrip() {
    export TEST_DIR=/tmp/mimixbox/it/cpio
    cd ${TEST_DIR}/in
    printf 'file.txt\n' | cpio -o > ${TEST_DIR}/arch.cpio
    mkdir -p ${TEST_DIR}/out && cd ${TEST_DIR}/out
    cpio -i < ${TEST_DIR}/arch.cpio
    printf '%s' "$(< ${TEST_DIR}/out/file.txt)"
}
TestCpioList() {
    export TEST_DIR=/tmp/mimixbox/it/cpio
    cd ${TEST_DIR}/in
    printf 'file.txt\n' | cpio -o > ${TEST_DIR}/arch2.cpio
    cpio -i -t < ${TEST_DIR}/arch2.cpio
}
