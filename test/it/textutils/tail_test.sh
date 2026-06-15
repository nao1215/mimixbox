Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    export TEST_FILE=${MIMIXBOX_IT_ROOT}/tail.txt
    mkdir -p ${TEST_DIR}
    printf '1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n' > ${TEST_FILE}
}

CleanUp() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    rm -rf ${TEST_DIR}
}

TestTailDefault() {
    export TEST_FILE=${MIMIXBOX_IT_ROOT}/tail.txt
    tail ${TEST_FILE}
}

TestTailLines() {
    export TEST_FILE=${MIMIXBOX_IT_ROOT}/tail.txt
    tail -n 3 ${TEST_FILE}
}

TestTailBytes() {
    printf 'hello world' | tail -c 5
}

TestTailPipe() {
    printf 'a\nb\nc\nd\n' | tail -n 2
}

TestTailNoExistFile() {
    tail /no_exist_file
}

TestTailFollow() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    export FOLLOW_FILE=${MIMIXBOX_IT_ROOT}/follow.txt
    mkdir -p "${TEST_DIR}"
    printf 'start\n' > "${FOLLOW_FILE}"
    # Append more data shortly after tail starts following.
    ( sleep 0.2; printf 'appended\n' >> "${FOLLOW_FILE}" ) &
    # tail -f follows forever; bound it with timeout (exits non-zero when killed),
    # so the assertion is on the captured output.
    timeout 0.5 tail -f -s 0.05 "${FOLLOW_FILE}"
}
