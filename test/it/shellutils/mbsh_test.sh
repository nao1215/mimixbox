Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/mbsh
    mkdir -p ${TEST_DIR}
}

CleanUp() {
    rm -rf ${MIMIXBOX_IT_ROOT}/mbsh
}

TestMbshEcho() {
    printf 'echo hello\nexit\n' | mbsh 2>/dev/null
}

TestMbshComment() {
    printf '# a comment\necho ok\nexit\n' | mbsh 2>/dev/null
}

TestMbshLastStatus() {
    printf 'false\necho status=$?\nexit\n' | mbsh 2>/dev/null
}

TestMbshCd() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/mbsh
    printf 'cd %s\npwd\nexit\n' "${TEST_DIR}" | mbsh 2>/dev/null
}

TestMbshCatConsumesInput() {
    printf 'cat\nhello\nexit\n' | mbsh 2>/dev/null
}

TestMbshNoReparse() {
    err=$(printf 'cat\nhello\nexit\n' | mbsh 2>&1 >/dev/null)
    case "${err}" in
        *"not a mimixbox command"*) echo reparsed ;;
        *) echo ok ;;
    esac
}

TestMbshDoubleQuote() {
    printf 'echo "a b"\nexit\n' | mbsh 2>/dev/null
}

TestMbshVarExpand() {
    out=$(printf 'echo $HOME\nexit\n' | mbsh 2>/dev/null)
    case "${out}" in
        *"${HOME}"*) echo expanded ;;
        *) echo literal ;;
    esac
}

TestMbshEnvAssignment() {
    printf 'FOO=bar env\nexit\n' | mbsh 2>/dev/null | grep '^FOO=bar$'
}

TestMbshSequence() {
    d=$(mktemp -d)
    printf 'echo one > %s/o; echo two >> %s/o\nexit\n' "$d" "$d" | mbsh >/dev/null 2>&1
    cat "$d/o"
    rm -rf "$d"
}

TestMbshPipeline() {
    d=$(mktemp -d)
    printf 'printf foo | wc -c > %s/o\nexit\n' "$d" | mbsh >/dev/null 2>&1
    tr -d ' \n' < "$d/o"
    rm -rf "$d"
}

TestMbshRedirectIn() {
    d=$(mktemp -d)
    printf 'a\nb\nc\n' > "$d/in"
    printf 'wc -l < %s/in > %s/o\nexit\n' "$d" "$d" | mbsh >/dev/null 2>&1
    tr -d ' \n' < "$d/o"
    rm -rf "$d"
}
