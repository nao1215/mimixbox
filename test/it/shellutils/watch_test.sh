TestWatchOnce() {
    # watch loops forever; bound it with timeout. It exits non-zero (killed),
    # so the assertion is on the captured output, which contains the command's.
    timeout 0.6 watch -t -n 0.2 echo tick 2>/dev/null
}
