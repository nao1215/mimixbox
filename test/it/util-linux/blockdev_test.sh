# Reading a block device requires privilege, so the e2e exercises the --help
# path and the no-query error; the ioctl dispatch is covered by Go unit tests.
TestBlockdevHelp() { blockdev --help; }
TestBlockdevNoQuery() { blockdev /dev/null 2>/dev/null; echo "rc=$?"; }
