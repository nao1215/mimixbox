Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/mktemp
    mkdir -p ${TEST_DIR}
}

CleanUp() {
    rm -rf ${MIMIXBOX_IT_ROOT}/mktemp
}

TestMktempFile() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/mktemp
    f=$(mktemp -p ${TEST_DIR})
    test -f "$f" && echo "created"
}

TestMktempDir() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/mktemp
    d=$(mktemp -d -p ${TEST_DIR})
    test -d "$d" && echo "created"
}

TestMktempDryRun() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/mktemp
    f=$(mktemp -u -p ${TEST_DIR})
    test ! -e "$f" && echo "not created"
}
