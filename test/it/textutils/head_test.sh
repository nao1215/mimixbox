Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    export TEST_FILE=${MIMIXBOX_IT_ROOT}/head.txt
    mkdir -p ${TEST_DIR}
    printf '1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n' > ${TEST_FILE}
}

CleanUp() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    rm -rf ${TEST_DIR}
}

TestHeadDefault() {
    export TEST_FILE=${MIMIXBOX_IT_ROOT}/head.txt
    head ${TEST_FILE}
}

TestHeadLines() {
    export TEST_FILE=${MIMIXBOX_IT_ROOT}/head.txt
    head -n 3 ${TEST_FILE}
}

TestHeadBytes() {
    printf 'hello world' | head -c 5
}

TestHeadPipe() {
    printf 'a\nb\nc\nd\n' | head -n 2
}

TestHeadNoExistFile() {
    head /no_exist_file
}
