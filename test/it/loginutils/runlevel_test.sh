# The CI/WSL utmp has no run-level record, so runlevel prints 'unknown' and
# exits 1 there; the parsing of a real record is covered by Go unit tests.
TestRunlevelRuns() { runlevel; echo "rc=$?"; }
