TestEnvAssign() {
    env FOO=bar | grep '^FOO=bar$'
}

TestEnvIgnore() {
    env -i ONLY=here
}

TestEnvRunCommand() {
    env GREETING=hi sh -c 'echo $GREETING'
}
