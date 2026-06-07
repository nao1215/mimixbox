Setup() {
    export TEST_DIR=/tmp/mimixbox/it
    export TEST_FILE=/tmp/mimixbox/it/expand.txt
    mkdir -p ${TEST_DIR}
    printf 'a\tb\n' > ${TEST_FILE}
}

CleanUp() {
    export TEST_DIR=/tmp/mimixbox/it
    rm -rf ${TEST_DIR}
}

TestExpandPipe() {
    printf 'a\tb\n' | expand
}

TestExpandTabStop() {
    printf 'a\tb\n' | expand -t 4
}

TestExpandFile() {
    export TEST_FILE=/tmp/mimixbox/it/expand.txt
    expand ${TEST_FILE}
}

TestExpandNoExistFile() {
    expand /no_exist_file
}
