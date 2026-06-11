TestMke2fsMagic() {
    dd if=/dev/zero of="$TEST_DIR/e.img" bs=1024 count=1024 2>/dev/null
    mke2fs "$TEST_DIR/e.img" >/dev/null
    # ext2 magic 0xEF53 (little-endian 53 ef) at superblock offset 1024+56 = 1080.
    dd if="$TEST_DIR/e.img" bs=1 skip=1080 count=2 2>/dev/null | od -An -tx1 | tr -d ' \n'
}
TestMkfsExt2TooLarge() {
    dd if=/dev/zero of="$TEST_DIR/b.img" bs=1024 count=10000 2>/dev/null
    mkfs.ext2 "$TEST_DIR/b.img" 2>/dev/null
    echo "rc=$?"
}
