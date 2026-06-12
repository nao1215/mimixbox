# Deallocating a VT needs a console and privilege; the e2e exercises the
# deterministic invalid-arg path. The ioctl dispatch is covered by unit tests.
TestDeallocvtBadN() { deallocvt notanumber 2>/dev/null; echo "rc=$?"; }
TestDeallocvtHelp() { deallocvt --help; }
