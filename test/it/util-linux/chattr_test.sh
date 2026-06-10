# Setting immutable/append needs privilege and ext2/4; the e2e exercises the
# deterministic mode-validation path. The flag math is covered by unit tests.
TestChattrBadMode() { chattr xi /tmp/f 2>/dev/null; echo "rc=$?"; }
TestChattrBadAttr() { chattr +Z /tmp/f 2>/dev/null; echo "rc=$?"; }
