Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    export TEST_FILE_NL=${MIMIXBOX_IT_ROOT}/nl.txt
    export TEST_FILE_NL2=${MIMIXBOX_IT_ROOT}/nl2.txt
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
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    rm  -rf ${TEST_DIR} 
}

TestNlNoArg() {
    export TEST_FILE_NL=${MIMIXBOX_IT_ROOT}/nl.txt
    nl ${TEST_FILE_NL}
}

TestNlFromPipeData() {
    export TEST_FILE_NL=${MIMIXBOX_IT_ROOT}/nl.txt
    echo ${TEST_FILE_NL} | nl
}

TestNlOnlyOperandWithPipeData() {
    export TEST_FILE_NL=${MIMIXBOX_IT_ROOT}/nl.txt
    echo ${TEST_FILE_NL} | nl ${TEST_FILE_NL}
}

TestNlConcatenateTwoFile() {
    export TEST_FILE_NL=${MIMIXBOX_IT_ROOT}/nl.txt
    export TEST_FILE_NL2=${MIMIXBOX_IT_ROOT}/nl2.txt
    nl ${TEST_FILE_NL} ${TEST_FILE_NL2}
}

TestNlHeredoc() {
    export TEST_FILE_NL=${MIMIXBOX_IT_ROOT}/nl.txt
    export TEST_FILE_NL2=${MIMIXBOX_IT_ROOT}/nl2.txt
    nl - << EOS ${TEST_FILE_NL} > ${TEST_FILE_NL2}
fish
zsh
EOS
    cat ${TEST_FILE_NL2}
}

TestNlHeredocStatus() {
    export TEST_FILE_NL=${MIMIXBOX_IT_ROOT}/nl.txt
    export TEST_FILE_NL2=${MIMIXBOX_IT_ROOT}/nl2.txt
    nl - << EOS ${TEST_FILE_NL} > ${TEST_FILE_NL2}
fish
zsh
EOS
}

TestNlNoOperand() {
    nl no_exist_file
}

SetupNlSections() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    export TEST_FILE_NL_SEC=${MIMIXBOX_IT_ROOT}/nl_sec.txt
    export TEST_FILE_NL_BLANK=${MIMIXBOX_IT_ROOT}/nl_blank.txt
    export LANG=C

    mkdir -p ${TEST_DIR}

    # Header/body/footer fixture using nl's delimiter lines.
    {
        printf 'H1\n'
        printf '\\:\\:\\:\n'
        printf 'HDR\n'
        printf '\\:\\:\n'
        printf 'B1\n'
        printf '\\:\n'
        printf 'F1\n'
    } > ${TEST_FILE_NL_SEC}

    # Blank-line fixture: a, four blank lines, b.
    printf 'a\n\n\n\n\nb\n' > ${TEST_FILE_NL_BLANK}
}

TestNlSectionsAll() {
    export TEST_FILE_NL_SEC=${MIMIXBOX_IT_ROOT}/nl_sec.txt
    nl -h a -b a -f a ${TEST_FILE_NL_SEC}
}

TestNlSectionsMixed() {
    export TEST_FILE_NL_SEC=${MIMIXBOX_IT_ROOT}/nl_sec.txt
    nl -h a -b t -f n ${TEST_FILE_NL_SEC}
}

TestNlJoinBlankLines() {
    export TEST_FILE_NL_BLANK=${MIMIXBOX_IT_ROOT}/nl_blank.txt
    nl -b a -l 2 ${TEST_FILE_NL_BLANK}
}