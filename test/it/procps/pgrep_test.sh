# pgrep for a long-running helper started by the test.
TestPgrepFindsSleep() {
    sleep 30 &
    p=$!
    found=$(pgrep sleep | grep -c "^${p}$")
    kill $p 2>/dev/null
    echo "$found"
}

TestPgrepNoMatch() {
    pgrep zzz_no_such_proc_zzz
    echo "rc=$?"
}
