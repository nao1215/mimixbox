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

TestViArrowNoAppend() {
    export TEST_DIR=/tmp/mimixbox/it/vi
    # Up-arrow (ESC [ A) must not be treated as 'A' (append); the file with the
    # following '!' must stay unchanged after :wq.
    printf '\033[A!\033:wq\r' | vi ${TEST_DIR}/a.txt
    printf '%s' "$(< ${TEST_DIR}/a.txt)"
}

TestViYankPaste() {
    export TEST_DIR=/tmp/mimixbox/it/vi
    printf 'one\ntwo\n' > "${TEST_DIR}/yp.txt"
    printf 'yyp:wq\r' | vi "${TEST_DIR}/yp.txt"
    printf '%s' "$(< ${TEST_DIR}/yp.txt)"
}

TestViCountedDelete() {
    export TEST_DIR=/tmp/mimixbox/it/vi
    printf 'abcdef\n' > "${TEST_DIR}/cd.txt"
    printf '2x:wq\r' | vi "${TEST_DIR}/cd.txt"
    printf '%s' "$(< ${TEST_DIR}/cd.txt)"
}

TestViUndo() {
    export TEST_DIR=/tmp/mimixbox/it/vi
    printf 'keepme\n' > "${TEST_DIR}/u.txt"
    printf 'xu:wq\r' | vi "${TEST_DIR}/u.txt"
    printf '%s' "$(< ${TEST_DIR}/u.txt)"
}

TestViSearchNext() {
    export TEST_DIR=/tmp/mimixbox/it/vi
    printf 'x\nfoo\ny\nfoo\nz\n' > "${TEST_DIR}/sn.txt"
    # /foo lands on the first match (line 2); n moves to the next (line 4); dd
    # deletes that line.
    printf '/foo\rndd:wq\r' | vi "${TEST_DIR}/sn.txt"
    printf '%s' "$(< ${TEST_DIR}/sn.txt)"
}
