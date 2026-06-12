# runsvdir supervises in the foreground until interrupted, so the e2e exercises
# the no-directory path and --help; the per-service start is covered by unit tests.
TestRunsvdirNoDir() { runsvdir 2>/dev/null; echo "rc=$?"; }
TestRunsvdirHelp() { runsvdir --help; }
