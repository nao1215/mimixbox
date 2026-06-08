Setup() {
    export TEST_DIR=/tmp/mimixbox/it
    mkdir -p ${TEST_DIR}
    printf 'alpha\nsecret\nbeta\n' > ${TEST_DIR}/words.txt
}
CleanUp() { rm -rf /tmp/mimixbox/it; }
TestPwcrack() {
    pwcrack -w /tmp/mimixbox/it/words.txt '$6$abcdefgh$ltjgWl6579NluT/Vi1nwEvcil.G5Nbc4NiXZaNGStk8PSwGfQv72N2CKPPrVACtLtip/cZ/1GM/O6IND4WQhG.'
}
