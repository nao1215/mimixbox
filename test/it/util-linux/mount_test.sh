TestMountListsRoot() { mount | grep -cE ' on / type '; }
TestMountRejectsMount() { mount /dev/sda1 /mnt 2>/dev/null; echo "rc=$?"; }
