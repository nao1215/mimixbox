TestFlockRuns() {
    flock "$TEST_DIR/lock" echo "locked-run"
}
TestFlockNonblock() {
    # Hold the lock with a background sleep, then a -n attempt must fail.
    flock "$TEST_DIR/lock" sleep 1 &
    sleep 0.2
    flock -n "$TEST_DIR/lock" echo "nope" 2>/dev/null
    echo "rc=$?"
    wait
}
