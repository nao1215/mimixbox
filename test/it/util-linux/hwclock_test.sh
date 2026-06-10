# Reading /dev/rtc requires privilege, so the e2e exercises the --help path; the
# RTC read and time formatting are covered by Go unit tests.
TestHwclockHelp() { hwclock --help; }
