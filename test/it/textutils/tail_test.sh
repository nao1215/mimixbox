Setup() {
    export TEST_DIR=/tmp/mimixbox/it
    export TEST_FILE=/tmp/mimixbox/it/tail.txt
    mkdir -p ${TEST_DIR}
    printf '1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n' > ${TEST_FILE}
}

CleanUp() {
    export TEST_DIR=/tmp/mimixbox/it
    rm -rf ${TEST_DIR}
}

TestTailDefault() {
    export TEST_FILE=/tmp/mimixbox/it/tail.txt
    tail ${TEST_FILE}
}

TestTailLines() {
    export TEST_FILE=/tmp/mimixbox/it/tail.txt
    tail -n 3 ${TEST_FILE}
}

TestTailBytes() {
    printf 'hello world' | tail -c 5
}

TestTailPipe() {
    printf 'a\nb\nc\nd\n' | tail -n 2
}

TestTailNoExistFile() {
    tail /no_exist_file
}
