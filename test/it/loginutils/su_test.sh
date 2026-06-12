# Switching users needs privilege and real auth, so the e2e exercises the
# deterministic unknown-user path; the auth/dispatch flow is unit-tested.
TestSuUnknownUser() { echo "" | su nonexistent_user_xyz 2>/dev/null; echo "rc=$?"; }
TestSuHelp() { su --help; }
