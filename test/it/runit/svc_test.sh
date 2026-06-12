TestSvc() {
    d="$TEST_DIR/svc"; mkdir -p "$d/supervise"; : > "$d/supervise/control"
    svc -d "$d"
    cat "$d/supervise/control"
}
TestSvcNoCmd() { d="$TEST_DIR/s2"; mkdir -p "$d/supervise"; : > "$d/supervise/control"; svc "$d" 2>/dev/null; echo "rc=$?"; }
