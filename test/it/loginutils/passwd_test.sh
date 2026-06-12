# Writing /etc/shadow needs privilege; the e2e exercises the deterministic
# conflicting-flags path. The change/lock/unlock/delete flows are unit-tested.
TestPasswdConflict() { passwd -l -u alice 2>/dev/null; echo "rc=$?"; }
TestPasswdHelp() { passwd --help; }
