# Real rfkill devices vary by host; the e2e confirms 'list' runs cleanly and the
# error paths. The sysfs parsing is covered by Go unit tests.
TestRfkillList() { rfkill list >/dev/null; echo "rc=$?"; }
TestRfkillUnknown() { rfkill bogus 2>/dev/null; echo "rc=$?"; }
