# crontab uses the system spool dir (privileged); the e2e exercises the
# deterministic unsupported-edit path. Install/list/remove are unit-tested.
TestCrontabEdit() { crontab -e 2>/dev/null; echo "rc=$?"; }
TestCrontabHelp() { crontab --help; }
