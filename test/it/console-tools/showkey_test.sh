# showkey needs a console in raw mode; the e2e exercises the deterministic
# validation and capability paths. The mode selection is covered by unit tests.
TestShowkeyConflict() { showkey -a -s 2>/dev/null; echo "rc=$?"; }
# Without a real console, showkey fails deterministically rather than hanging.
TestShowkeyNoConsole() { showkey </dev/null 2>/dev/null; echo "rc=$?"; }
