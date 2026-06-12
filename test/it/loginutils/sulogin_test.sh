# sulogin authenticates root and starts a root shell (privileged); the e2e
# exercises the wrong-password path and --help. The flow is unit-tested.
TestSuloginWrongPw() { echo "definitely_wrong_password" | sulogin 2>/dev/null; echo "rc=$?"; }
TestSuloginHelp() { sulogin --help; }
