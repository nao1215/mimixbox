TestGettyPrompt() {
    # Empty username -> getty fails without invoking login, but still prints the prompt.
    printf '\n' | getty tty1 2>/dev/null | grep -c 'login: '
}
TestGettyNoTTY() { printf 'alice\n' | getty 2>/dev/null; echo "rc=$?"; }
