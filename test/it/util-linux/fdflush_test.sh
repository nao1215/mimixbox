TestFdflushNoDev() { fdflush 2>/dev/null; echo "rc=$?"; }
TestFdflushHelp() { fdflush --help; }
