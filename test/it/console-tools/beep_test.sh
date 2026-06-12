# The console speaker needs a console and privilege; the e2e exercises the
# deterministic invalid-option paths. The dispatch is covered by unit tests.
TestBeepBadFreq() { beep -f 0 2>/dev/null; echo "rc=$?"; }
TestBeepBadRepeat() { beep -r 0 2>/dev/null; echo "rc=$?"; }
