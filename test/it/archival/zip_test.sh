Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/zip
    mkdir -p ${TEST_DIR}
    printf 'zipped' > ${TEST_DIR}/a.txt
}
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}/zip; }

TestZipThenUnzipList() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/zip
    cd ${TEST_DIR}
    zip out.zip a.txt >/dev/null
    unzip -l out.zip
}
TestZipThenExtract() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/zip
    cd ${TEST_DIR}
    zip out2.zip a.txt >/dev/null
    unzip -d ${TEST_DIR}/dest out2.zip >/dev/null
    printf '%s' "$(< ${TEST_DIR}/dest/a.txt)"
}
