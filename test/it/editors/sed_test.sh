TestSedSubstitute() {
    printf 'hello world\n' | sed 's/world/sed/'
}
TestSedGlobal() {
    printf 'a a a\n' | sed 's/a/b/g'
}
TestSedDelete() {
    printf '1\n2\n3\n' | sed '2d'
}
TestSedPrintN() {
    printf 'x\ny\nz\n' | sed -n '2p'
}
