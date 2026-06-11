TestSsdStartStop() {
    p="$TEST_DIR/foo.pid"
    start-stop-daemon -S -p "$p" -x /bin/sleep -- 30 >/dev/null 2>&1
    pid=$(cat "$p")
    start-stop-daemon -K -p "$p" >/dev/null 2>&1
    # After stopping, the process must be gone and the pidfile removed.
    if kill -0 "$pid" 2>/dev/null; then echo alive; else echo "stopped"; fi
}
TestSsdNoMode() { start-stop-daemon -p /tmp/x 2>/dev/null; echo "rc=$?"; }
