# who reads the binary utmp database. To keep the test independent of the host's
# login sessions, it reads an empty utmp (/dev/null), which has no entries.

TestWhoEmpty() {
    who /dev/null
    echo "rc=$?"
}

TestWhoCount() {
    who -q /dev/null
}

TestWhoHelp() {
    who --help
}
