# runsv supervises in the foreground until interrupted, so the e2e exercises the
# no-directory path and --help; the restart loop is covered by Go unit tests.
TestRunsvNoDir() { runsv 2>/dev/null; echo "rc=$?"; }
TestRunsvHelp() { runsv --help; }
