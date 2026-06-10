# Operating a CD-ROM needs a real drive and privilege, so the e2e exercises the
# --help and missing-device paths; the ioctl dispatch is covered by unit tests.
TestEjectHelp() { eject --help; }
TestEjectMissing() { eject /dev/no_such_cdrom 2>/dev/null; echo "rc=$?"; }
