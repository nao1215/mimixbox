# A Linux virtual console isn't available on CI/WSL, so the e2e exercises the
# deterministic failure path and --help; the VT read is covered by unit tests.
TestFgconsoleNoVT() { fgconsole </dev/null 2>/dev/null; echo "rc=$?"; }
TestFgconsoleHelp() { fgconsole --help; }
