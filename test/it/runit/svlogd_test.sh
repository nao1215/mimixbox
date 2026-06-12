TestSvlogd() {
    printf 'hello\nworld\n' | svlogd "$TEST_DIR/log"
    cat "$TEST_DIR/log/current"
}
TestSvlogdNoDir() { echo x | svlogd 2>/dev/null; echo "rc=$?"; }
