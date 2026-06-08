TestStringsPipe() {
    printf 'hi\000hello\000world' | strings
}
