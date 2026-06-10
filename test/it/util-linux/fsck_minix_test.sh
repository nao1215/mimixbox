TestFsckMinixRoundTrip() {
    dd if=/dev/zero of="$TEST_DIR/f.img" bs=1024 count=2048 2>/dev/null
    mkfs.minix "$TEST_DIR/f.img" >/dev/null
    fsck.minix "$TEST_DIR/f.img" | sed -n '1p'
}
TestFsckMinixBad() {
    dd if=/dev/zero of="$TEST_DIR/b.img" bs=1024 count=4 2>/dev/null
    fsck.minix "$TEST_DIR/b.img" 2>/dev/null
    echo "rc=$?"
}
