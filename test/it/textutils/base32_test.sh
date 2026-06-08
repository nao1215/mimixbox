TestBase32EncodePipe() {
    printf 'hello\n' | base32
}
TestBase32DecodePipe() {
    printf 'NBSWY3DPBI======\n' | base32 -d
}
