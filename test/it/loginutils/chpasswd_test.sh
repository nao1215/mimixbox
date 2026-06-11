# chpasswd writes /etc/shadow (privileged); the e2e exercises the deterministic
# unknown-method path. The shadow update is covered by Go unit tests.
TestChpasswdBadMethod() { echo 'alice:secret' | chpasswd -c bogus 2>/dev/null; echo "rc=$?"; }
TestChpasswdHelp() { chpasswd --help; }
