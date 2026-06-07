TestOdChar() {
    printf 'ABC\n' | od -c
}

TestOdHex() {
    printf 'AB' | od -A x -t x1
}

TestOdNoAddr() {
    printf 'A' | od -A n -t o1
}
