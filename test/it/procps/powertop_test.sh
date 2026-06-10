# Power-supply hardware is absent on CI/WSL, so the report path is covered by Go
# unit tests with fixtures; the e2e confirms powertop runs and exits 0.
TestPowertopRuns() { powertop >/dev/null; echo "rc=$?"; }
TestPowertopHelp() { powertop --help; }
