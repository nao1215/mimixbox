Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/grep_gnu
    mkdir -p ${TEST_DIR}/src ${TEST_DIR}/vendor
    printf '1\n2\nMATCH\nb\nc\n' > ${TEST_DIR}/ctx.txt
    printf 'a\nMATCH\nb\n' > ${TEST_DIR}/ctx2.txt
    printf 'MATCH\nb\nc\nd\ne\nf\nMATCH\n' > ${TEST_DIR}/groups.txt
    printf 'needle\n' > ${TEST_DIR}/keep.go
    printf 'needle\n' > ${TEST_DIR}/skip.txt
    printf 'needle\n' > ${TEST_DIR}/app.log
    printf 'needle\n' > ${TEST_DIR}/src/a.txt
    printf 'needle\n' > ${TEST_DIR}/vendor/b.txt
    printf 'aaa\nbbb\nccc\n' > ${TEST_DIR}/off.txt
    printf 'needle\n' > ${TEST_DIR}/hit.txt
    printf 'nothing\n' > ${TEST_DIR}/miss.txt
}
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}/grep_gnu; }

TestGrepAfter() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/grep_gnu
    grep -A1 MATCH ${TEST_DIR}/ctx.txt
}
TestGrepBefore() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/grep_gnu
    grep -B1 MATCH ${TEST_DIR}/ctx2.txt
}
TestGrepContext() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/grep_gnu
    grep -C1 MATCH ${TEST_DIR}/ctx2.txt
}
TestGrepGroupSeparator() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/grep_gnu
    grep -A1 MATCH ${TEST_DIR}/groups.txt
}
TestGrepInclude() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/grep_gnu
    grep -r --include='*.go' needle ${TEST_DIR}
}
TestGrepExclude() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/grep_gnu
    grep -r --exclude='*.log' needle ${TEST_DIR}
}
TestGrepExcludeDir() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/grep_gnu
    grep -r --exclude-dir=vendor needle ${TEST_DIR}
}
TestGrepColor() {
    grep --color=always world <<'EOF'
hello world
EOF
}
TestGrepByteOffset() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/grep_gnu
    grep -b bbb ${TEST_DIR}/off.txt
}
TestGrepFilesWithoutMatch() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/grep_gnu
    grep -L needle ${TEST_DIR}/hit.txt ${TEST_DIR}/miss.txt
}
