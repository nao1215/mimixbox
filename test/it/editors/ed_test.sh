Setup() {
    export TEST_DIR=/tmp/mimixbox/it/ed
    mkdir -p ${TEST_DIR}
    printf 'one\ntwo\nthree\n' > ${TEST_DIR}/buf.txt
}

CleanUp() { rm -rf /tmp/mimixbox/it/ed; }

TestEdPrint() {
    printf '1,$p\nq\n' | ed ${TEST_DIR}/buf.txt
}

TestEdAppendWrite() {
    printf '2a\nINSERTED\n.\nw\nq\n' | ed ${TEST_DIR}/buf.txt > /dev/null
    cat ${TEST_DIR}/buf.txt
}

TestEdSubstitute() {
    printf '2s/two/TWO/\nw\nq\n' | ed ${TEST_DIR}/buf.txt > /dev/null
    cat ${TEST_DIR}/buf.txt
}
