TestFsckMinix() {
    dd if=/dev/zero of="$TEST_DIR/m.img" bs=1024 count=2048 2>/dev/null
    mkfs.minix "$TEST_DIR/m.img" >/dev/null
    fsck "$TEST_DIR/m.img"
}
TestFsckUnknown() {
    dd if=/dev/zero of="$TEST_DIR/u.img" bs=1024 count=8 2>/dev/null
    fsck "$TEST_DIR/u.img" 2>/dev/null
    echo "rc=$?"
}
