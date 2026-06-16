Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/tee_output_error
    mkdir -p ${TEST_DIR}
}

CleanUp() {
    rm -rf ${MIMIXBOX_IT_ROOT}/tee_output_error
}

# A writable destination with an explicit MODE succeeds and copies the input.
TestTeeOutputErrorWritable() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/tee_output_error
    printf 'hello\n' | tee --output-error=warn ${TEST_DIR}/good.txt > /dev/null
    cat ${TEST_DIR}/good.txt
}

# warn mode keeps writing to the good file even though the bad path fails, and
# exits nonzero. We print the good file's content so the spec can assert it was
# still written; the exit status reflects the failure.
TestTeeOutputErrorWarnContinues() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/tee_output_error
    printf 'payload\n' | tee --output-error=warn \
        ${TEST_DIR}/missing-dir/bad.txt ${TEST_DIR}/good.txt > /dev/null 2>&1
    local rc=$?
    cat ${TEST_DIR}/good.txt 2>/dev/null
    return ${rc}
}

# exit mode stops at the first failing destination, so the good file that comes
# after the bad one is never created. We emit "absent" when it does not exist.
TestTeeOutputErrorExitStops() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/tee_output_error
    printf 'payload\n' | tee --output-error=exit \
        ${TEST_DIR}/missing-dir/bad.txt ${TEST_DIR}/good.txt > /dev/null 2>&1
    local rc=$?
    if [ -f ${TEST_DIR}/good.txt ]; then
        printf 'present\n'
    else
        printf 'absent\n'
    fi
    return ${rc}
}
