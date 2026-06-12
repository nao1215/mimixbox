# run-init needs to run as PID 1 with privilege, so the e2e exercises the
# deterministic validation paths; the switch and exec are unit-tested.
TestRunInitMissingInit() { run-init /tmp 2>/dev/null; echo "rc=$?"; }
TestRunInitBadDir() { run-init /no/such/dir /init 2>/dev/null; echo "rc=$?"; }
