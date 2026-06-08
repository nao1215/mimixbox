Setup() {
    export TEST_DIR=/tmp/mimixbox/it
    mkdir -p ${TEST_DIR}
    printf 'x' > ${TEST_DIR}/rl_target
    ln -sf /tmp/mimixbox/it/rl_target /tmp/mimixbox/it/rl_link
}
CleanUp() { rm -rf /tmp/mimixbox/it; }
TestReadlink() {
    readlink /tmp/mimixbox/it/rl_link
}
