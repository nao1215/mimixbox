# Writing the account databases needs privilege; the e2e exercises the
# deterministic no-argument path. The account creation is covered by unit tests.
TestAdduserNoName() { adduser 2>/dev/null; echo "rc=$?"; }
TestAdduserHelp() { adduser --help; }
