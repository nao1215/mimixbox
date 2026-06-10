# The test's own process has its cwd here, so fuser . finds at least one PID.
TestFuserCwd() {
    fuser . 2>/dev/null | grep -cE '[0-9]'
}

TestFuserNone() {
    fuser /no/such/fuser/file 2>/dev/null
    echo "rc=$?"
}
