Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/mv_gnu
    mkdir -p ${TEST_DIR}/dst
    printf 'A\n' > ${TEST_DIR}/a.txt
    printf 'B\n' > ${TEST_DIR}/b.txt
}

CleanUp() {
    rm -rf ${MIMIXBOX_IT_ROOT}/mv_gnu
}

# mv --target-directory dst a b moves each source into dst (destination-first).
TestMvTargetDirectory() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/mv_gnu
    mv --target-directory ${TEST_DIR}/dst ${TEST_DIR}/a.txt ${TEST_DIR}/b.txt
    ls ${TEST_DIR}/dst
}

# mv --update keeps a destination that is newer than the source.
TestMvUpdateKeepsNewer() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/mv_gnu
    # Create the source first, then the destination, so the destination's
    # modification time is strictly newer than the source's.
    printf 'source\n' > ${TEST_DIR}/src.txt
    sleep 1
    printf 'newer-dest\n' > ${TEST_DIR}/dest.txt
    mv --update ${TEST_DIR}/src.txt ${TEST_DIR}/dest.txt
    cat ${TEST_DIR}/dest.txt
}
