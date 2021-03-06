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
