TestUniqBasic() {
    printf 'a\na\nb\nc\nc\nc\n' | uniq
}

TestUniqCount() {
    printf 'a\na\nb\nc\nc\nc\n' | uniq -c
}

TestUniqRepeated() {
    printf 'a\na\nb\nc\nc\n' | uniq -d
}

TestUniqUnique() {
    printf 'a\na\nb\nc\nc\n' | uniq -u
}
