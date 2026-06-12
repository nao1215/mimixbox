TestSvokSupervised() {
    mkdir -p "$TEST_DIR/svc/supervise"
    : > "$TEST_DIR/svc/supervise/ok"
    svok "$TEST_DIR/svc"; echo "rc=$?"
}
TestSvokNotSupervised() {
    mkdir -p "$TEST_DIR/down"
    svok "$TEST_DIR/down"; echo "rc=$?"
}
