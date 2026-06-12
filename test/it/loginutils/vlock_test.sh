# vlock verifies the current user's password against the system backend, which
# fails unprivileged; the e2e exercises the wrong-password path and --help. The
# unlock flow is unit-tested.
TestVlockWrongPw() { echo "definitely_wrong_xyz" | vlock >/dev/null 2>&1; echo "rc=$?"; }
TestVlockHelp() { vlock --help; }
