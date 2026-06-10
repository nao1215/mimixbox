# Trimming needs privilege, so the e2e exercises the deterministic
# missing-argument path; the ioctl dispatch is covered by unit tests.
TestFstrimNoArg() { fstrim 2>/dev/null; echo "rc=$?"; }
TestFstrimHelp() { fstrim --help; }
