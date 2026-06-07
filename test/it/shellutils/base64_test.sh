Setup() {
    export TEST_DIR=/tmp/mimixbox/it
    export TEST_FILE=/tmp/mimixbox/it/base64.txt
    mkdir -p ${TEST_DIR}
    printf 'hello\n' > ${TEST_FILE}
}

CleanUp() {
    export TEST_DIR=/tmp/mimixbox/it
    rm -rf ${TEST_DIR}
}

TestBase64EncodePipe() {
    printf 'hello\n' | base64
}

TestBase64EncodeFile() {
    export TEST_FILE=/tmp/mimixbox/it/base64.txt
    base64 ${TEST_FILE}
}

TestBase64DecodePipe() {
    printf 'aGVsbG8K\n' | base64 -d
}

TestBase64RoundTrip() {
    printf 'MimixBox\n' | base64 | base64 -d
}

TestBase64NoExistFile() {
    base64 /no_exist_file
}
