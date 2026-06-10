# Freeing a ramdisk needs the device and privilege; the e2e exercises the
# deterministic no-device path. The ioctl dispatch is covered by unit tests.
TestFreeramdiskNoDev() { freeramdisk 2>/dev/null; echo "rc=$?"; }
TestFreeramdiskHelp() { freeramdisk --help; }
