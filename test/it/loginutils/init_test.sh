TestInitRunsActions() {
    printf 'si::sysinit:echo SYSINIT\nl::wait:echo WAIT\n' > "$TEST_DIR/inittab"
    init -t "$TEST_DIR/inittab"
}
TestInitMissing() { init -t "$TEST_DIR/no_such_inittab" 2>/dev/null; echo "rc=$?"; }
