TestDeluserNoName() { deluser 2>/dev/null; echo "rc=$?"; }
TestDeluserHelp() { deluser --help; }
