Setup() { export D=${MIMIXBOX_IT_ROOT}; mkdir -p $D; printf 'hello gzip roundtrip\n' > $D/g.txt; }
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}; }
TestGzipRoundtrip() {
    gzip ${MIMIXBOX_IT_ROOT}/g.txt
    gunzip ${MIMIXBOX_IT_ROOT}/g.txt.gz
    cat ${MIMIXBOX_IT_ROOT}/g.txt
}
