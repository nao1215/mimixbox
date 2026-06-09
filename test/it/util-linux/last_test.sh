# An empty wtmp (here /dev/null) means no history: last exits 0 with no output.
TestLastRuns() {
    out=$(last /dev/null)
    echo "[$out] rc=$?"
}

TestLastMissing() {
    last /no/such/mimixbox/wtmp 2>/dev/null
    echo "rc=$?"
}
