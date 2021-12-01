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

TestEchoPipeWithoutXargs() {
    echo "pipe" | echo
}

TestEchoPipeWithargs() {
    echo "pipe" | xargs echo
}

TestEchoNoArg() {
    echo
}
