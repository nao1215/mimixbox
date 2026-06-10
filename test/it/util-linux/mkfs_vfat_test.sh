TestMkfsVfatSig() {
    dd if=/dev/zero of="$TEST_DIR/v.img" bs=1024 count=8192 2>/dev/null
    mkfs.vfat "$TEST_DIR/v.img" >/dev/null
    # FAT16 fs-type label at byte offset 54.
    dd if="$TEST_DIR/v.img" bs=1 skip=54 count=5 2>/dev/null
}
TestMkdosfsTooSmall() {
    dd if=/dev/zero of="$TEST_DIR/s.img" bs=1024 count=512 2>/dev/null
    mkdosfs "$TEST_DIR/s.img" 2>/dev/null
    echo "rc=$?"
}
