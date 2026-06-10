# Build a tiny MBR with one Linux partition using printf, then list it.
TestFdiskList() {
    img="$TEST_DIR/disk.img"
    dd if=/dev/zero of="$img" bs=512 count=1 2>/dev/null
    # type byte 0x83 (Linux) at offset 450 (446+4)
    printf '\203' | dd of="$img" bs=1 seek=450 count=1 conv=notrunc 2>/dev/null
    # LBA start = 2048 (0x00000800 little-endian) at offset 454
    printf '\000\010\000\000' | dd of="$img" bs=1 seek=454 count=4 conv=notrunc 2>/dev/null
    # sector count = 100 (0x64) at offset 458
    printf '\144\000\000\000' | dd of="$img" bs=1 seek=458 count=4 conv=notrunc 2>/dev/null
    # MBR signature 0x55AA at 510
    printf '\125\252' | dd of="$img" bs=1 seek=510 count=2 conv=notrunc 2>/dev/null
    fdisk -l "$img" | grep -c 'Linux'
}
TestFdiskBad() {
    dd if=/dev/zero of="$TEST_DIR/n.img" bs=512 count=1 2>/dev/null
    fdisk -l "$TEST_DIR/n.img" 2>/dev/null
    echo "rc=$?"
}
