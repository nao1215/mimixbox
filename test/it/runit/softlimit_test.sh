# mbsh has no ulimit builtin, so the e2e confirms softlimit runs a program under
# an open-file limit (exit 0); the rlimit values are checked by Go unit tests.
TestSoftlimitRuns() { softlimit -o 64 true; echo "rc=$?"; }
TestSoftlimitNoProg() { softlimit -o 64 2>/dev/null; echo "rc=$?"; }
