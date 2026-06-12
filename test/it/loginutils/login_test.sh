# login authenticates and starts a shell with the user's credentials
# (privileged); the e2e exercises the unknown-user path and --help. The flow is
# covered by Go unit tests.
TestLoginUnknownUser() { printf 'nope\n' | login nonexistent_user_xyz 2>/dev/null; echo "rc=$?"; }
TestLoginHelp() { login --help; }
