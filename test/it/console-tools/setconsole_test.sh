# Redirecting the console needs privilege; the e2e exercises a missing-device
# failure and --help. The ioctl dispatch is covered by unit tests.
TestSetconsoleBadDev() { setconsole /dev/no_such_console 2>/dev/null; echo "rc=$?"; }
TestSetconsoleHelp() { setconsole --help; }
