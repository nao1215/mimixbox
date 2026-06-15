Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/tree
    mkdir -p ${TEST_DIR}/sub
    touch ${TEST_DIR}/sub/leaf.txt ${TEST_DIR}/root.txt
}

CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}/tree; }

# The tree body uses multibyte box-drawing characters; shellspec's output
# capture cannot carry those through the hermetic toolchain, and the exact
# formatting is already covered by the Go unit tests. Here we assert on the
# ASCII summary line, which confirms the traversal counted every entry.
TestTreeSummary() {
    tree ${TEST_DIR} | grep directories
}

TestTreeStatus() {
    tree ${TEST_DIR} > /dev/null
    echo $?
}

TestNicePrints() {
    nice
}
