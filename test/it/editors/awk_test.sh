TestAwkField() {
    printf 'one two three\n' | awk '{print $2}'
}
TestAwkFS() {
    printf 'root:x:0\n' | awk -F: '{print $1}'
}
TestAwkNR() {
    printf 'a\nb\nc\n' | awk 'NR==2'
}
TestAwkEnd() {
    printf 'a\nb\nc\n' | awk 'END{print NR}'
}
