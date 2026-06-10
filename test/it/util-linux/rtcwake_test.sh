TestRtcwakeSuspendRejected() { rtcwake -m mem -s 10 2>/dev/null; echo "rc=$?"; }
TestRtcwakeNoTime() { rtcwake -m no 2>/dev/null; echo "rc=$?"; }
