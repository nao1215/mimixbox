# Most namespaces need privilege, so the e2e exercises the deterministic
# no-namespace-flag path; the flag math and command exec are unit-tested.
TestUnshareNoFlag() { unshare echo x 2>/dev/null; echo "rc=$?"; }
TestUnshareHelp() { unshare --help; }
