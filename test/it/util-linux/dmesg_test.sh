# Reading the kernel ring buffer can require privilege depending on the host's
# dmesg_restrict, so the e2e exercises the sysfs-independent --help path; the
# read and priority-stripping are covered by Go unit tests.
TestDmesgHelp() { dmesg --help; }
