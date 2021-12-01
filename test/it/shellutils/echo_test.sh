TestEchoNormal() {
    mimixbox echo "Hello World!"
}

TestEchoVariable() {
    mimixbox echo "Hello $1"
}

TestEchoEnvVariable() {
    export TEST_ENV="TEST_ENV_VAR"
    mimixbox echo ${TEST_ENV}
}

TestEchoPipeWithoutXargs() {
    mimixbox echo "pipe" | mimixbox echo
}

TestEchoPipeWithargs() {
    mimixbox echo "pipe" | xargs mimixbox echo
}

TestEchoNoArg() {
    mimixbox echo
}
