Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    mkdir -p ${TEST_DIR}
    printf 'alice:x:1000:1000:Alice:/home/alice:/bin/sh\n' > ${TEST_DIR}/passwd
    printf 'alice:$6$abc$HASH:19000:0:99999:7:::\n' > ${TEST_DIR}/shadow
}
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}; }
TestUnshadow() {
    unshadow ${MIMIXBOX_IT_ROOT}/passwd ${MIMIXBOX_IT_ROOT}/shadow
}
