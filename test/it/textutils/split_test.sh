Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    mkdir -p ${TEST_DIR}
}
CleanUp() {
    rm -rf ${MIMIXBOX_IT_ROOT}
}
TestSplitLines() {
    printf '1\n2\n3\n' | split -l 2 - ${MIMIXBOX_IT_ROOT}/part-
    cat ${MIMIXBOX_IT_ROOT}/part-aa
}
TestSplitNumeric() {
    printf '1\n2\n3\n4\n5\n' | split -l 2 -d - ${MIMIXBOX_IT_ROOT}/num-
    ls ${MIMIXBOX_IT_ROOT}/num-* | sed "s#${MIMIXBOX_IT_ROOT}/##" | sort | tr '\n' ' ' | sed 's/ $//'
}
TestSplitNumericContent() {
    printf '1\n2\n3\n4\n5\n' | split -l 2 -d - ${MIMIXBOX_IT_ROOT}/num-
    cat ${MIMIXBOX_IT_ROOT}/num-00
}
TestSplitAdditionalSuffix() {
    printf '1\n2\n3\n' | split -l 2 --additional-suffix=.txt - ${MIMIXBOX_IT_ROOT}/add-
    ls ${MIMIXBOX_IT_ROOT}/add-* | sed "s#${MIMIXBOX_IT_ROOT}/##" | sort | tr '\n' ' ' | sed 's/ $//'
}
TestSplitSuffixLength() {
    printf '1\n2\n' | split -l 1 -a 3 - ${MIMIXBOX_IT_ROOT}/len-
    ls ${MIMIXBOX_IT_ROOT}/len-* | sed "s#${MIMIXBOX_IT_ROOT}/##" | sort | tr '\n' ' ' | sed 's/ $//'
}
