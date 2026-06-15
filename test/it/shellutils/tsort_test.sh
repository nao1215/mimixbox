# shellcheck shell=sh
# Integration helper for `tsort` (topological sort). See test/it/README.md.

Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/tsort
    export LANG=C
    mkdir -p "${TEST_DIR}"
    # Fixture: dependency-graph edges "a depends-before b".
    {
        printf 'a b\n'
        printf 'b c\n'
        printf 'c d\n'
    } > "${TEST_DIR}/graph.txt"
}

CleanUp() { rm -rf "${MIMIXBOX_IT_ROOT}/tsort"; }

# Read edges from a file fixture and emit a valid topological order.
TestTsortFromFile() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/tsort
    tsort "${TEST_DIR}/graph.txt"
}

# Read edges from stdin.
TestTsortFromStdin() {
    printf 'a b\nb c\n' | tsort
}
