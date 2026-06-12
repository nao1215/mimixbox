TestEnvdir() {
    mkdir -p "$TEST_DIR/env"
    printf 'hello\n' > "$TEST_DIR/env/GREETING"
    envdir "$TEST_DIR/env" sh -c 'echo "$GREETING"'
}
TestEnvdirNoArgs() { envdir /only 2>/dev/null; echo "rc=$?"; }
