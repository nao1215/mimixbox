TestSetlogconsBadN() { setlogcons notanumber 2>/dev/null; echo "rc=$?"; }
TestSetlogconsHelp() { setlogcons --help; }
