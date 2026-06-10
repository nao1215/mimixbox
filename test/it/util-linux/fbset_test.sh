# No framebuffer device is available on CI, so the e2e exercises the deterministic
# missing-device path and --help; the mode formatting is covered by unit tests.
TestFbsetNoFb() { fbset -fb /dev/no_such_fb 2>/dev/null; echo "rc=$?"; }
TestFbsetHelp() { fbset --help; }
