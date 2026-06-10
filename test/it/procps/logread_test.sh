TestLogreadFile() {
    printf 'msg one\nmsg two\n' > "$TEST_DIR/sys.log"
    logread "$TEST_DIR/sys.log"
}
TestLogreadMissing() {
    logread "$TEST_DIR/nope.log" 2>/dev/null
    echo "rc=$?"
}
