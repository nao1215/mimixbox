# Entering namespaces needs privilege, so the e2e exercises the deterministic
# validation paths; the setns dispatch and command exec are unit-tested.
TestNsenterNoTarget() { nsenter -n echo x 2>/dev/null; echo "rc=$?"; }
TestNsenterNoNs() { nsenter -t 1 echo x 2>/dev/null; echo "rc=$?"; }
