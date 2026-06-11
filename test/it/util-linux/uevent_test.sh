# uevent blocks monitoring the netlink socket (privileged), so the e2e exercises
# --help; the receive/parse loop is covered by Go unit tests.
TestUeventHelp() { uevent --help; }
