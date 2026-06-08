Setup() {
    export TEST_DIR=/tmp/mimixbox/it/compress
    mkdir -p ${TEST_DIR}
    printf 'compress me compress me compress me\n' > ${TEST_DIR}/data.txt
}
CleanUp() { rm -rf /tmp/mimixbox/it/compress; }

TestCompressRoundTrip() {
    export TEST_DIR=/tmp/mimixbox/it/compress
    cp ${TEST_DIR}/data.txt ${TEST_DIR}/rt.txt
    compress ${TEST_DIR}/rt.txt
    uncompress ${TEST_DIR}/rt.txt.Z
    printf '%s' "$(< ${TEST_DIR}/rt.txt)"
}
