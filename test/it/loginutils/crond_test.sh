# crond -f runs until interrupted (and fires only on minute boundaries), so the
# e2e exercises the no-foreground path and --help; the cron-matching engine is
# covered by Go unit tests.
TestCrondNoForeground() { crond 2>/dev/null; echo "rc=$?"; }
TestCrondHelp() { crond --help; }
