# Switching VTs needs a console and privilege; the e2e exercises the deterministic
# validation paths. The ioctl dispatch is covered by unit tests.
TestChvtBadN() { chvt notanumber 2>/dev/null; echo "rc=$?"; }
TestChvtNoN() { chvt 2>/dev/null; echo "rc=$?"; }
