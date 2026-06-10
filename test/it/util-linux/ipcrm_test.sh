TestIpcrmNothing() {
    ipcrm 2>/dev/null
    echo "rc=$?"
}

TestIpcrmBadId() {
    ipcrm -q 2147483647 2>/dev/null
    echo "rc=$?"
}
