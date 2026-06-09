TestHd() {
    printf 'hello world\n' | hd
}

TestHexdumpCanonical() {
    printf 'hello world\n' | hexdump -C
}

TestHexdumpDefault() {
    printf 'hello world\n' | hexdump
}

TestGetopt() {
    getopt -o ab: --long alpha,beta: -- -a -b val --alpha pos
}

TestGetoptEval() {
    eval set -- "$(getopt -o n: -- -n hello world)"
    printf '%s|' "$@"
}
