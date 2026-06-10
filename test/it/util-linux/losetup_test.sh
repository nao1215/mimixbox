# Active loop devices vary by host; the e2e confirms -a runs cleanly (exit 0)
# and that a setup request is refused. Listing logic is covered by unit tests.
TestLosetupAll() { losetup -a >/dev/null; echo "rc=$?"; }
TestLosetupSetup() { losetup /dev/loop0 /tmp/img 2>/dev/null; echo "rc=$?"; }
