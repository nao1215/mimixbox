# pivot_root needs privilege and a prepared mount; the e2e exercises the
# deterministic argument-validation path. The dispatch is covered by unit tests.
TestPivotRootBadArgs() { pivot_root /onlyone 2>/dev/null; echo "rc=$?"; }
TestPivotRootHelp() { pivot_root --help; }
