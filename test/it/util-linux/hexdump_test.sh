TestHd() {
    printf 'hello world\n' | hd
}

TestHexdumpCanonical() {
    printf 'hello world\n' | hexdump -C
}

TestHexdumpDefault() {
    printf 'hello world\n' | hexdump
}
