TestBootchartdSample() {
    bootchartd -o "$TEST_DIR/bc" >/dev/null
    # A proc_stat.log with the CPU line must have been written.
    grep -c '^cpu ' "$TEST_DIR/bc/proc_stat.log"
}
