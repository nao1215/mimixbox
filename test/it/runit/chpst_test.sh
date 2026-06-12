TestChpstEnvdir() {
    mkdir -p "$TEST_DIR/env"
    printf 'world\n' > "$TEST_DIR/env/HELLO"
    chpst -e "$TEST_DIR/env" sh -c 'echo "$HELLO"'
}
TestChpstNoProg() { chpst -o 64 2>/dev/null; echo "rc=$?"; }
