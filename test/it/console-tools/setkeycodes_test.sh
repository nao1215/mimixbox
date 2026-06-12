TestSetkeycodesOdd() { setkeycodes e060 2>/dev/null; echo "rc=$?"; }
TestSetkeycodesBad() { setkeycodes zz 1 2>/dev/null; echo "rc=$?"; }
