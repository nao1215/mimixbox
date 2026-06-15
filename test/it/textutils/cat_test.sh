Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    export TEST_FILE_CAT=${MIMIXBOX_IT_ROOT}/cat.txt
    export TEST_FILE_CAT2=${MIMIXBOX_IT_ROOT}/cat2.txt
    export LANG=C

    mkdir -p ${TEST_DIR}

    echo "sh" > ${TEST_FILE_CAT}
    echo "ash" >> ${TEST_FILE_CAT}
    echo "csh" >> ${TEST_FILE_CAT}
    echo "bash" >> ${TEST_FILE_CAT}

   echo "fish" > ${TEST_FILE_CAT2}
   echo "zsh" >> ${TEST_FILE_CAT2}
}

CleanUp() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    rm  -rf ${TEST_DIR} 
}

TestCatNoArg() {
    export TEST_FILE_CAT=${MIMIXBOX_IT_ROOT}/cat.txt
    cat ${TEST_FILE_CAT}
}

TestCatWithNumbetOption() {
    export TEST_FILE_CAT=${MIMIXBOX_IT_ROOT}/cat.txt
    cat -n ${TEST_FILE_CAT}
}

TestCatFromPipeData() {
    export TEST_FILE_CAT=${MIMIXBOX_IT_ROOT}/cat.txt
    echo ${TEST_FILE_CAT} | cat
}

TestCatOnlyOperandWithPipeData() {
    export TEST_FILE_CAT=${MIMIXBOX_IT_ROOT}/cat.txt
    export TEST_FILE_CAT2=${MIMIXBOX_IT_ROOT}/cat2.txt
    echo ${TEST_FILE_CAT2} | cat ${TEST_FILE_CAT}
}

TestCatConcatenateTwoFile() {
    export TEST_FILE_CAT=${MIMIXBOX_IT_ROOT}/cat.txt
    export TEST_FILE_CAT2=${MIMIXBOX_IT_ROOT}/cat2.txt
    cat ${TEST_FILE_CAT} ${TEST_FILE_CAT2}
}

TestCatConcatenateTwoFileWithNumberOption() {
    export TEST_FILE_CAT=${MIMIXBOX_IT_ROOT}/cat.txt
    export TEST_FILE_CAT2=${MIMIXBOX_IT_ROOT}/cat2.txt
    cat -n ${TEST_FILE_CAT} ${TEST_FILE_CAT2}
}

TestCatHeredoc() {
    export TEST_FILE_CAT=${MIMIXBOX_IT_ROOT}/cat.txt
    export TEST_FILE_CAT2=${MIMIXBOX_IT_ROOT}/cat2.txt
    cat - << EOS ${TEST_FILE_CAT} > ${TEST_FILE_CAT2}
fish
zsh
EOS
    cat ${TEST_FILE_CAT2}
}

TestCatNoOperand() {
    cat no_exist_file
}