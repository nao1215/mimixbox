# acpid -f reads /proc/acpi/event (often absent) and blocks; the e2e exercises
# the no-foreground path and --help. Event dispatch is covered by Go unit tests.
TestAcpidNoForeground() { acpid 2>/dev/null; echo "rc=$?"; }
TestAcpidHelp() { acpid --help; }
