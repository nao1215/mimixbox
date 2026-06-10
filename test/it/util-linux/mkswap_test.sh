TestMkswapImage() {
    dd if=/dev/zero of="$TEST_DIR/swap.img" bs=1024 count=64 2>/dev/null
    chmod 0600 "$TEST_DIR/swap.img"
    mkswap "$TEST_DIR/swap.img" | grep -c 'version 1'
}
TestMkswapSignature() {
    dd if=/dev/zero of="$TEST_DIR/swap2.img" bs=1024 count=64 2>/dev/null
    chmod 0600 "$TEST_DIR/swap2.img"
    mkswap "$TEST_DIR/swap2.img" >/dev/null
    # The signature SWAPSPACE2 must be present in the formatted image.
    od -c "$TEST_DIR/swap2.img" | grep -c 'S   W   A   P   S   P   A   C   E   2'
}
