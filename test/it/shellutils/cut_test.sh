TestCutField() {
    printf 'a,b,c\n' | cut -f 2 -d ,
}

TestCutFields() {
    printf 'a,b,c,d\n' | cut -f 1,3 -d ,
}

TestCutFieldRange() {
    printf 'a,b,c,d\n' | cut -f 2- -d ,
}

TestCutChars() {
    printf 'abcdef\n' | cut -c 1-3
}

TestCutNoList() {
    printf 'a,b\n' | cut -d ,
}
