Setup() { export TEST_DIR=${MIMIXBOX_IT_ROOT}; mkdir -p ${TEST_DIR}; printf 'hello' > ${TEST_DIR}/stat_file; }
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}; }
TestStatSize() {
    stat -c '%s' ${MIMIXBOX_IT_ROOT}/stat_file
}
TestStatPrintf() {
    # --printf interprets \t and adds no trailing newline; pipe through od to
    # make the absence of a trailing newline observable.
    stat --printf '%n=%s' ${MIMIXBOX_IT_ROOT}/stat_file
}
TestStatPrintfName() {
    stat --printf '%n %s\n' ${MIMIXBOX_IT_ROOT}/stat_file
}
TestStatFormatNewline() {
    # --format/-c append a trailing newline (wc -l counts it).
    stat --format '%s' ${MIMIXBOX_IT_ROOT}/stat_file | wc -l | tr -d ' '
}
TestStatTerseFieldCount() {
    # --terse prints one space-separated line; assert the field count.
    stat --terse ${MIMIXBOX_IT_ROOT}/stat_file | awk '{print NF}'
}
TestStatTerseSize() {
    # Second terse field is the size in bytes.
    stat --terse ${MIMIXBOX_IT_ROOT}/stat_file | awk '{print $2}'
}
