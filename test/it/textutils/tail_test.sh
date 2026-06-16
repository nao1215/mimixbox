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

# TestTailFollowPid follows a file while watching a short-lived process. tail
# must exit on its own promptly after that process dies. A generous timeout
# guards the test: if --pid is broken, tail would run until the timeout fires
# (exit 124), so a success exit proves --pid stopped the follow loop itself.
TestTailFollowPid() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    export FOLLOW_FILE=${MIMIXBOX_IT_ROOT}/follow_pid.txt
    mkdir -p "${TEST_DIR}"
    printf 'start\n' > "${FOLLOW_FILE}"
    # A ~1s sleeper stands in for the process being waited on.
    sleep 1 &
    sleeper_pid=$!
    # Append once while the sleeper is still alive so output is observable.
    ( sleep 0.2; printf 'appended\n' >> "${FOLLOW_FILE}" ) &
    # If --pid works, tail exits shortly after the sleeper ends (~1s), well
    # before the 5s timeout. timeout returns 124 only if tail keeps running.
    timeout 5 tail -f -s 0.1 --pid="${sleeper_pid}" "${FOLLOW_FILE}"
}
