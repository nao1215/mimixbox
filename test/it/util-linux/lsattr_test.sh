# FS_IOC_GETFLAGS support depends on the underlying filesystem (tmpfs lacks it),
# so the e2e exercises --help; the attribute decoding is covered by unit tests.
TestLsattrHelp() { lsattr --help; }
