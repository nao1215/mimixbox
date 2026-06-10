# A FAT filesystem isn't available on CI, so the e2e exercises the deterministic
# validation paths; the attribute math is covered by unit tests.
TestFatattrNoFile() { fatattr 2>/dev/null; echo "rc=$?"; }
TestFatattrBadAttr() { fatattr +Z /tmp/x 2>/dev/null; echo "rc=$?"; }
