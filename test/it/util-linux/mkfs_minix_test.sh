TestMkfsMinixMagic() {
    dd if=/dev/zero of="$TEST_DIR/m.img" bs=1024 count=2048 2>/dev/null
    mkfs.minix "$TEST_DIR/m.img" >/dev/null
    # The Minix v1 magic 0x137F (little-endian 7f 13) sits at byte offset 1040.
    dd if="$TEST_DIR/m.img" bs=1 skip=1040 count=2 2>/dev/null | od -An -tx1 | tr -d ' \n'
}
TestMkfsMinixTooSmall() {
    dd if=/dev/zero of="$TEST_DIR/s.img" bs=1024 count=4 2>/dev/null
    mkfs.minix "$TEST_DIR/s.img" 2>/dev/null
    echo "rc=$?"
}
