# switch_root needs to run as PID 1 with privilege, so the e2e exercises the
# deterministic validation paths; the switch and exec are covered by unit tests.
TestSwitchRootMissingInit() { switch_root /tmp 2>/dev/null; echo "rc=$?"; }
TestSwitchRootBadDir() { switch_root /no/such/dir /init 2>/dev/null; echo "rc=$?"; }
