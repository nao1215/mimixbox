# chsh rewrites /etc/passwd (privileged); the e2e exercises the deterministic
# read-only and rejection paths that need no write access. The passwd update is
# covered by Go unit tests.
TestChshHelp() { chsh --help; }
TestChshListShells() { chsh -l >/dev/null; echo "rc=$?"; }
# A non-existent user can never be changed, so this fails regardless of privilege.
TestChshUnknownUser() { chsh -s /bin/sh mimixbox-no-such-user-xyz 2>/dev/null; echo "rc=$?"; }
# A relative shell path is rejected before any write is attempted.
TestChshRelativeShell() { chsh -s relative/shell root 2>/dev/null; echo "rc=$?"; }
