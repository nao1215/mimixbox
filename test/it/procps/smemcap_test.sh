# Capture to a tar and confirm the meminfo entry is present.
TestSmemcap() {
    smemcap > "$TEST_DIR/cap.tar"
    tar -tf "$TEST_DIR/cap.tar" | grep -c '^meminfo$'
}
