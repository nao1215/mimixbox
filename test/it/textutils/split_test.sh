Setup() {
    export TEST_DIR=/tmp/mimixbox/it
    mkdir -p ${TEST_DIR}
}
CleanUp() {
    rm -rf /tmp/mimixbox/it
}
TestSplitLines() {
    printf '1\n2\n3\n' | split -l 2 - /tmp/mimixbox/it/part-
    cat /tmp/mimixbox/it/part-aa
}
