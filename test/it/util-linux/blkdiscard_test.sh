# Discarding needs a real block device and privilege; the e2e exercises the
# deterministic no-device path. The range math is covered by unit tests.
TestBlkdiscardNoDev() { blkdiscard 2>/dev/null; echo "rc=$?"; }
TestBlkdiscardHelp() { blkdiscard --help; }
