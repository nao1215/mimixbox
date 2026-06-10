# With no terminals recorded for the harness, wall writes nothing and exits 0.
# The banner formatting and targeting are covered by the Go unit tests.
TestWallRuns() {
    echo "broadcast" | wall
    echo "rc=$?"
}
