# Fetching from a live RFC 868 server is non-deterministic, so the e2e exercises
# the error paths; the fetch/print/set logic is covered by Go unit tests.
TestRdateNoHost() {
    rdate 2>/dev/null
    echo "rc=$?"
}

TestRdateUnreachable() {
    rdate 127.0.0.1 2>/dev/null
    echo "rc=$?"
}
