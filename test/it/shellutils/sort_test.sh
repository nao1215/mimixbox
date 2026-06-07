TestSortLexical() {
    printf 'banana\napple\ncherry\n' | sort
}

TestSortNumeric() {
    printf '10\n2\n1\n' | sort -n
}

TestSortReverse() {
    printf 'a\nb\nc\n' | sort -r
}

TestSortUnique() {
    printf 'a\na\nb\n' | sort -u
}
