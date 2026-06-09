# Note: bash's own builtins shadow "[" and "test", so the bracket aliases are
# invoked explicitly through the mimixbox dispatcher.
TestBracketTrue() {
    mimixbox '[' -f /etc/hosts ']' && echo yes || echo no
}

TestBracketFalse() {
    mimixbox '[' -f /no/such/mimixbox/file ']' && echo yes || echo no
}

TestBusyboxDispatch() {
    d=$(mktemp -d)
    printf 'hello\n' > "$d/f"
    busybox cat "$d/f"
    rm -rf "$d"
}

TestBusyboxList() {
    busybox --list
}

TestShDashC() {
    sh -c 'echo from-sh'
}

TestBashStdinNoPrompt() {
    out=$(printf 'echo via-bash\n' | bash 2>/dev/null)
    case "$out" in
        *"mbsh:"*) echo prompted ;;
        *via-bash*) echo ok ;;
        *) echo "$out" ;;
    esac
}
