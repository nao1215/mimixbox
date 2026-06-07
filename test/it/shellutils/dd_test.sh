TestDdCopy() {
    printf 'hello world' | dd status=none
}

TestDdCount() {
    printf 'hello world' | dd bs=1 count=5 status=none
}

TestDdUcase() {
    printf 'abc' | dd conv=ucase status=none
}
