Setup() {
    export TEST_DIR=/tmp/mimixbox/it
    export TEST_FILE=/tmp/mimixbox/it/tac.txt
    mkdir -p ${TEST_DIR}
    printf 'first\nsecond\nthird\n' > ${TEST_FILE}
}

CleanUp() {
    export TEST_DIR=/tmp/mimixbox/it
    rm -rf ${TEST_DIR}
}

TestTacFile() {
    export TEST_FILE=/tmp/mimixbox/it/tac.txt
    tac ${TEST_FILE}
}

TestTacPipe() {
    printf 'a\nb\nc\n' | tac
}

TestTacNoExistFile() {
    tac /no_exist_file
}
