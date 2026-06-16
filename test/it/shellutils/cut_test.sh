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

TestCutComplementFields() {
    printf 'a,b,c\n' | cut -f 2 -d , --complement
}

TestCutComplementBytes() {
    printf 'abcde\n' | cut -b 2-3 --complement
}

# TestCutZeroTerminatedFields cuts the second field of each NUL-terminated
# record and renders the NUL boundaries as '|' so the result is comparable as
# plain text.
TestCutZeroTerminatedFields() {
    printf 'a,b,c\000d,e,f\000' | cut -f 2 -d , -z | tr '\000' '|'
}

# TestCutZeroTerminatedBytes keeps the first two bytes of each NUL record.
TestCutZeroTerminatedBytes() {
    printf 'abc\000def\000' | cut -b 1-2 -z | tr '\000' '|'
}
