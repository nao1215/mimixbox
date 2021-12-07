Setup() {
    export TEST_DIR=/tmp/mimixbox/it
    export TEST_FILE_CAT=/tmp/mimixbox/it/cat.txt
    export TEST_FILE_CAT2=/tmp/mimixbox/it/cat2.txt
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
    export TEST_DIR=/tmp/mimixbox/it
    rm  -rf ${TEST_DIR} 
}

TestCatNoArg() {
    export TEST_FILE_CAT=/tmp/mimixbox/it/cat.txt
    cat ${TEST_FILE_CAT}
}

TestCatWithNumbetOption() {
    export TEST_FILE_CAT=/tmp/mimixbox/it/cat.txt
    cat -n ${TEST_FILE_CAT}
}

TestCatFromPipeData() {
    export TEST_FILE_CAT=/tmp/mimixbox/it/cat.txt
    echo ${TEST_FILE_CAT} | cat
}

TestCatOnlyOperandWithPipeData() {
    export TEST_FILE_CAT=/tmp/mimixbox/it/cat.txt
    echo ${TEST_FILE_CAT} | cat ${TEST_FILE_CAT}
}

TestCatConcatenateTwoFile() {
    export TEST_FILE_CAT=/tmp/mimixbox/it/cat.txt
    export TEST_FILE_CAT2=/tmp/mimixbox/it/cat2.txt
    cat ${TEST_FILE_CAT} ${TEST_FILE_CAT2}
}

TestCatConcatenateTwoFileWithNumberOption() {
    export TEST_FILE_CAT=/tmp/mimixbox/it/cat.txt
    export TEST_FILE_CAT2=/tmp/mimixbox/it/cat2.txt
    cat -n ${TEST_FILE_CAT} ${TEST_FILE_CAT2}
}

TestCatHeredoc() {
    export TEST_FILE_CAT=/tmp/mimixbox/it/cat.txt
    export TEST_FILE_CAT2=/tmp/mimixbox/it/cat2.txt
    cat - << EOS ${TEST_FILE_CAT} > ${TEST_FILE_CAT2}
fish
zsh
EOS
    cat ${TEST_FILE_CAT2}
}

TestCatNoOperand() {
    cat no_exist_file
}