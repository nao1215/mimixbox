# Dropping to another user's uid/gid needs privilege; the e2e exercises the
# deterministic unknown-user and arg-count paths. The dispatch is unit-tested.
TestSetuidgidUnknown() { setuidgid nonexistent_user_xyz true 2>/dev/null; echo "rc=$?"; }
TestSetuidgidNoProg() { setuidgid root 2>/dev/null; echo "rc=$?"; }
