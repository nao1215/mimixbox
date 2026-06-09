Setup() {
    export TEST_DIR=/tmp/mimixbox/it/man
    mkdir -p ${TEST_DIR}/man1
    printf 'FOO(1)\nthe foo page\n' > ${TEST_DIR}/man1/foo.1
    printf 'BAR(1)\nthe bar page\n' | gzip > ${TEST_DIR}/man1/bar.1.gz
}

CleanUp() { rm -rf /tmp/mimixbox/it/man; }

TestManPlain() { man -M ${TEST_DIR} foo; }
TestManGzip() { man -M ${TEST_DIR} bar; }
TestManNotFound() {
    man -M ${TEST_DIR} missing 2>/dev/null
    echo "exit=$?"
}
