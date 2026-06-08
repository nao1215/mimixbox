TestEchoNormal() {
    echo "Hello World!"
}

TestEchoVariable() {
    echo "Hello $1"
}

TestEchoEnvVariable() {
    export TEST_ENV="TEST_ENV_VAR"
    echo ${TEST_ENV}
}

TestEchoPipeWithargs() {
    echo "pipe" | xargs echo
}

TestEchoNoArg() {
    echo
}

TestEchoRedirect() {
    echo "MimixBox" > /tmp/it/mimixbox/echo.txt
    cat /tmp/it/mimixbox/echo.txt
}

# echo is also a bash builtin, so the bare name would run the builtin instead of
# the MimixBox applet. type -P resolves the on-PATH executable (the MimixBox
# symlink) so these cases exercise the applet's --help/--version contract.
TestEchoHelp() {
    "$(type -P echo)" --help
}

TestEchoVersion() {
    "$(type -P echo)" --version
}

TestEchoHelpNotFirst() {
    "$(type -P echo)" foo --help
}
