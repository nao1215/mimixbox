TestFsfreezeNoMode() { fsfreeze /mnt 2>/dev/null; echo "rc=$?"; }
TestFsfreezeBothModes() { fsfreeze -f -u /mnt 2>/dev/null; echo "rc=$?"; }
