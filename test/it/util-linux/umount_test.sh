TestUmountNotMounted() { umount /not/a/real/mountpoint 2>/dev/null; echo "rc=$?"; }
TestUmountNoArg() { umount 2>/dev/null; echo "rc=$?"; }
