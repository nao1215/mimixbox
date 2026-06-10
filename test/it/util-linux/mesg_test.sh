# Under the e2e harness stdin is not a terminal, so mesg reports the error and
# exits 2. The terminal permission logic is covered by the Go unit tests.
TestMesgNotTty() {
    echo x | mesg 2>&1
    echo "rc=$?"
}
