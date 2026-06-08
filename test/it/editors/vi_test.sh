Setup() {
    export TEST_DIR=/tmp/mimixbox/it/vi
    mkdir -p ${TEST_DIR}
    printf 'hello\nworld\n' > ${TEST_DIR}/a.txt
    printf 'bar\n' > ${TEST_DIR}/b.txt
}
CleanUp() { rm -rf /tmp/mimixbox/it/vi; }

TestViDeleteChar() {
    export TEST_DIR=/tmp/mimixbox/it/vi
    printf 'x:wq\r' | vi ${TEST_DIR}/a.txt
    printf '%s' "$(< ${TEST_DIR}/a.txt)"
}

TestViInsert() {
    export TEST_DIR=/tmp/mimixbox/it/vi
    printf 'ifoo\033:wq\r' | vi ${TEST_DIR}/b.txt
    printf '%s' "$(< ${TEST_DIR}/b.txt)"
}

TestViNewFile() {
    export TEST_DIR=/tmp/mimixbox/it/vi
    printf 'icreated\033:wq\r' | vi ${TEST_DIR}/new.txt
    printf '%s' "$(< ${TEST_DIR}/new.txt)"
}
