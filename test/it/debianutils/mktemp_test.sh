Setup() {
    export TEST_DIR=/tmp/mimixbox/it/mktemp
    mkdir -p ${TEST_DIR}
}

CleanUp() {
    rm -rf /tmp/mimixbox/it/mktemp
}

TestMktempFile() {
    export TEST_DIR=/tmp/mimixbox/it/mktemp
    f=$(mktemp -p ${TEST_DIR})
    test -f "$f" && echo "created"
}

TestMktempDir() {
    export TEST_DIR=/tmp/mimixbox/it/mktemp
    d=$(mktemp -d -p ${TEST_DIR})
    test -d "$d" && echo "created"
}

TestMktempDryRun() {
    export TEST_DIR=/tmp/mimixbox/it/mktemp
    f=$(mktemp -u -p ${TEST_DIR})
    test ! -e "$f" && echo "not created"
}
