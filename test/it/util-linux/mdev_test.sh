# Creating /dev nodes needs privilege, so the e2e exercises the deterministic
# no-scan-flag path and --help; the scan/mknod logic is covered by unit tests.
TestMdevNoScan() { mdev 2>/dev/null; echo "rc=$?"; }
TestMdevHelp() { mdev --help; }
