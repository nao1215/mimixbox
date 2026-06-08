TestXargsEcho() {
    printf 'a b c\n' | xargs echo
}
TestXargsMaxArgs() {
    printf '1 2 3 4\n' | xargs -n 2 echo | wc -l | tr -d ' '
}
TestXargsReplace() {
    printf 'world\n' | xargs -I {} echo hello {}
}
