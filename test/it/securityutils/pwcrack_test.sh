Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    mkdir -p ${TEST_DIR}
    printf 'alpha\nsecret\nbeta\n' > ${TEST_DIR}/words.txt
}
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}; }
TestPwcrack() {
    pwcrack -w ${MIMIXBOX_IT_ROOT}/words.txt '$6$abcdefgh$ltjgWl6579NluT/Vi1nwEvcil.G5Nbc4NiXZaNGStk8PSwGfQv72N2CKPPrVACtLtip/cZ/1GM/O6IND4WQhG.'
}
