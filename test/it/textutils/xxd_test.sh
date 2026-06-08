TestXxdDump() {
    printf 'hello\n' | xxd
}
TestXxdRevert() {
    printf 'hello\n' | xxd | xxd -r
}
