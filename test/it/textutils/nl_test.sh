Setup() {
    export TEST_DIR=/tmp/mimixbox/it
    export TEST_FILE_NL=/tmp/mimixbox/it/nl.txt
    export TEST_FILE_NL2=/tmp/mimixbox/it/nl2.txt
    export LANG=C

    mkdir -p ${TEST_DIR}

    echo "sh" > ${TEST_FILE_NL}
    echo "ash" >> ${TEST_FILE_NL}
    echo "csh" >> ${TEST_FILE_NL}
    echo "bash" >> ${TEST_FILE_NL}

   echo "fish" > ${TEST_FILE_NL2}
   echo "zsh" >> ${TEST_FILE_NL2}
}

Cleanup() {
    export TEST_DIR=/tmp/mimixbox/it
    rm  -rf ${TEST_DIR} 
}

TestNlNoArg() {
    export TEST_FILE_NL=/tmp/mimixbox/it/nl.txt
    nl ${TEST_FILE_NL}
}

TestNlFromPipeData() {
    export TEST_FILE_NL=/tmp/mimixbox/it/nl.txt
    echo ${TEST_FILE_NL} | nl
}

TestNlOnlyOperandWithPipeData() {
    export TEST_FILE_NL=/tmp/mimixbox/it/nl.txt
    echo ${TEST_FILE_NL} | nl ${TEST_FILE_NL}
}

TestNlConcatenateTwoFile() {
    export TEST_FILE_NL=/tmp/mimixbox/it/nl.txt
    export TEST_FILE_NL2=/tmp/mimixbox/it/nl2.txt
    nl ${TEST_FILE_NL} ${TEST_FILE_NL2}
}

TestNlHeredoc() {
    export TEST_FILE_NL=/tmp/mimixbox/it/nl.txt
    export TEST_FILE_NL2=/tmp/mimixbox/it/nl2.txt
    nl - << EOS ${TEST_FILE_NL} > ${TEST_FILE_NL2}
fish
zsh
EOS
    cat ${TEST_FILE_NL2}
}

TestNlHeredocStatus() {
    export TEST_FILE_NL=/tmp/mimixbox/it/nl.txt
    export TEST_FILE_NL2=/tmp/mimixbox/it/nl2.txt
    nl - << EOS ${TEST_FILE_NL} > ${TEST_FILE_NL2}
fish
zsh
EOS
}

TestNlNoOperand() {
    nl no_exist_file
}