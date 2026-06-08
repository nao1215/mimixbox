TestPwgenCount() {
    pwgen -n 3 -l 8 | wc -l | tr -d ' '
}
