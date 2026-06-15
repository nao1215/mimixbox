Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    export TEST_FILE=${MIMIXBOX_IT_ROOT}/expand.txt
    mkdir -p ${TEST_DIR}
    printf 'a\tb\n' > ${TEST_FILE}
}

CleanUp() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    rm -rf ${TEST_DIR}
}

TestExpandPipe() {
    printf 'a\tb\n' | expand
}

TestExpandTabStop() {
    printf 'a\tb\n' | expand -t 4
}

TestExpandFile() {
    export TEST_FILE=${MIMIXBOX_IT_ROOT}/expand.txt
    expand ${TEST_FILE}
}

TestExpandNoExistFile() {
    expand /no_exist_file
}
