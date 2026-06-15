export TEST_DIR=${MIMIXBOX_IT_ROOT}/mv
export DIR_IN_TEST_DIR=${MIMIXBOX_IT_ROOT}/mv/inner
export TEST_DIR2=${MIMIXBOX_IT_ROOT}/mv2
export TEST_DIR3=${MIMIXBOX_IT_ROOT}/mv3
export TEST_DIR4=${MIMIXBOX_IT_ROOT}/mv4
export TEST_FILE=${MIMIXBOX_IT_ROOT}/mv/1.txt
export TEST_FILE2=${MIMIXBOX_IT_ROOT}/mv/2.txt
export TEST_FILE3=${MIMIXBOX_IT_ROOT}/mv/3.txt
export TEST_FILE_INNER=${DIR_IN_TEST_DIR}/inner.txt

Setup() {
    mkdir -p ${TEST_DIR}
    mkdir -p ${TEST_DIR2}
    mkdir -p ${TEST_DIR3}
    mkdir -p ${TEST_DIR4}
    mkdir -p ${DIR_IN_TEST_DIR}
    touch ${TEST_FILE}
    touch ${TEST_FILE2}
    touch ${TEST_FILE3}
    touch ${TEST_FILE_INNER}
}

Cleanup() {
    rm -rf "${MIMIXBOX_IT_ROOT}"
}

TestMvRename() {
    mv ${TEST_FILE} ${TEST_DIR}/rename.txt
    ls ${TEST_DIR}/rename.txt
}

TestMvRenameStatus() {
    mv ${TEST_FILE} ${TEST_DIR}/rename.txt
}

TestMvMoveFile() {
    mv ${TEST_FILE} ${DIR_IN_TEST_DIR}
    ls ${DIR_IN_TEST_DIR}
}

TestMvMoveFileStatus() {
    mv ${TEST_FILE} ${DIR_IN_TEST_DIR}
}

TestMvThreeFileAtSameTime() {
    mv ${TEST_FILE} ${TEST_FILE2} ${TEST_FILE3} ${DIR_IN_TEST_DIR}
    ls ${DIR_IN_TEST_DIR}
}

TestMvThreeFileAtSameTimeStatus() {
    mv ${TEST_FILE} ${TEST_FILE2} ${TEST_FILE3} ${DIR_IN_TEST_DIR}
}

TestMvThreeFileAndOneOfThreeFail() {
    mv ${TEST_FILE} ${TEST_DIR}/no_exist_file ${TEST_FILE3} ${DIR_IN_TEST_DIR}
    ls ${DIR_IN_TEST_DIR}
}

TestMvThreeFileAndOneOfThreeFailStatus() {
    mv ${TEST_FILE} ${TEST_DIR}/no_exist_file ${TEST_FILE3} ${DIR_IN_TEST_DIR}
}

TestMvDirToDir() {
    mv ${TEST_DIR2} ${TEST_DIR}
    ls ${TEST_DIR}
}

TestMvDirToDirStatus() {
    mv ${TEST_DIR2} ${TEST_DIR}
}

TestMvThreeDirs() {
    mv ${TEST_DIR2} ${TEST_DIR3} ${TEST_DIR4} ${TEST_DIR}
    ls ${TEST_DIR}
}

TestMvThreeStatus() {
    mv ${TEST_DIR2} ${TEST_DIR3} ${TEST_DIR4} ${TEST_DIR}
}

TestMvThreeDirsAndOneOfThreeFail() {
    mv ${TEST_DIR2} ${TEST_DIR}/no_exist_dir ${TEST_DIR4} ${DIR_IN_TEST_DIR}
    ls ${DIR_IN_TEST_DIR}
}

TestMvThreeDirsAndOneOfThreeFailStatus() {
    mv ${TEST_DIR2}  ${TEST_DIR}/no_exist_dir ${TEST_DIR4} ${DIR_IN_TEST_DIR}
}

TestMvFileAtSampePath() {
    mv  ${TEST_FILE} ${TEST_FILE}
}

TestMvSrcAndDestIsSameName() {
    touch ${TEST_DIR}/inner.txt
    mv ${TEST_DIR}/inner.txt ${TEST_FILE_INNER}
    ls ${TEST_DIR}
    ls ${DIR_IN_TEST_DIR}
}

TestMvSrcAndDestIsSameNameStatus() {
    touch ${TEST_DIR}/inner.txt
    mv ${TEST_DIR}/inner.txt ${TEST_FILE_INNER}
}

TestMvSrcAndDestIsSameNameWithBackupOpt() {
    touch ${TEST_DIR}/inner.txt
    mv -b ${TEST_DIR}/inner.txt ${DIR_IN_TEST_DIR}
    ls ${DIR_IN_TEST_DIR}
}

TestMvSrcAndDestIsSameNameWithBackupOptStatus() {
    touch ${TEST_DIR}/inner.txt
    mv -b ${TEST_DIR}/inner.txt ${DIR_IN_TEST_DIR}
}