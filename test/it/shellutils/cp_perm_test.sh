Setup() {
    export TEST_DIR=/tmp/mimixbox/it
    mkdir -p ${TEST_DIR}
    printf '#!/bin/sh\necho hi\n' > ${TEST_DIR}/script.sh
    chmod 755 ${TEST_DIR}/script.sh
    mkdir -m 700 ${TEST_DIR}/private
}
CleanUp() { chmod -R u+w /tmp/mimixbox/it 2>/dev/null; rm -rf /tmp/mimixbox/it; }
TestCpPreservesExecBit() {
    cp ${TEST_DIR}/script.sh ${TEST_DIR}/copy.sh
    stat -c '%a' ${TEST_DIR}/copy.sh
}
TestCpPreservesDirMode() {
    cp -r ${TEST_DIR}/private ${TEST_DIR}/private_copy
    stat -c '%a' ${TEST_DIR}/private_copy
}
TestCpForceOverwritesReadonly() {
    printf 'old\n' > ${TEST_DIR}/dst.txt
    chmod 444 ${TEST_DIR}/dst.txt
    printf 'new\n' > ${TEST_DIR}/src.txt
    cp -f ${TEST_DIR}/src.txt ${TEST_DIR}/dst.txt
    cat ${TEST_DIR}/dst.txt
}
