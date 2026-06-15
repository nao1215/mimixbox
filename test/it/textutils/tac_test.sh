Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    export TEST_FILE=${MIMIXBOX_IT_ROOT}/tac.txt
    mkdir -p ${TEST_DIR}
    printf 'first\nsecond\nthird\n' > ${TEST_FILE}
}

CleanUp() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    rm -rf ${TEST_DIR}
}

TestTacFile() {
    export TEST_FILE=${MIMIXBOX_IT_ROOT}/tac.txt
    tac ${TEST_FILE}
}

TestTacPipe() {
    printf 'a\nb\nc\n' | tac
}

TestTacNoExistFile() {
    tac /no_exist_file
}
