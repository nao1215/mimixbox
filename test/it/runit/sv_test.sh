TestSvUp() {
    d="$TEST_DIR/svc"; mkdir -p "$d/supervise"; : > "$d/supervise/control"; : > "$d/supervise/ok"
    sv up "$d"
    cat "$d/supervise/control"
}
TestSvStatus() {
    d="$TEST_DIR/svc2"; mkdir -p "$d/supervise"; : > "$d/supervise/ok"; echo 99 > "$d/supervise/pid"
    sv status "$d"
}
