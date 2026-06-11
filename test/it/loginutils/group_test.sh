# Writing /etc/group needs privilege; the e2e exercises the deterministic
# no-argument paths. The add/remove logic is covered by Go unit tests.
TestAddgroupNoName() { addgroup 2>/dev/null; echo "rc=$?"; }
TestDelgroupNoName() { delgroup 2>/dev/null; echo "rc=$?"; }
