TestSsdStartStop() {
    p="$TEST_DIR/foo.pid"
    start-stop-daemon -S -p "$p" -x /bin/sleep -- 30 >/dev/null 2>&1
    pid=$(cat "$p")
    start-stop-daemon -K -p "$p" >/dev/null 2>&1
    # After stopping, the process must be gone. SIGTERM delivery and reaping are
    # asynchronous, so poll briefly instead of checking once (avoids a flaky
    # "alive" result on slower hosts).
    i=0
    while [ "$i" -lt 50 ] && kill -0 "$pid" 2>/dev/null; do
        sleep 0.1
        i=$((i + 1))
    done
    if kill -0 "$pid" 2>/dev/null; then echo alive; else echo "stopped"; fi
}
TestSsdNoMode() { start-stop-daemon -p /tmp/x 2>/dev/null; echo "rc=$?"; }
