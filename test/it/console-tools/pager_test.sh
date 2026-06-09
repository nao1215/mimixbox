TestMorePassthrough() { printf 'a\nb\nc\n' | more; }
TestLessPassthrough() { printf 'x\ny\nz\n' | less; }
TestMoreFile() {
    d=$(mktemp -d); printf 'one\ntwo\n' > "$d/f"
    more "$d/f"
    rm -rf "$d"
}
