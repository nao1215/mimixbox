# /sys/bus/usb/devices may be absent on CI runners, so the e2e exercises the
# sysfs-independent --help path; the listing itself is covered by unit tests.
TestLsusbHelp() { lsusb --help; }
