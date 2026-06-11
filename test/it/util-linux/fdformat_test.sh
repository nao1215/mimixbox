# Formatting needs a real floppy drive and privilege, so the e2e exercises the
# deterministic no-device path and --help; the format sequence is unit-tested.
TestFdformatNoDev() { fdformat 2>/dev/null; echo "rc=$?"; }
TestFdformatHelp() { fdformat --help; }
