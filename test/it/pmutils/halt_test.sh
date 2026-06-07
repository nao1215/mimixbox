# NOTE: these tests intentionally only exercise --help / --version. Running halt,
# poweroff or reboot without those flags would actually stop the machine, so the
# real shutdown behaviour is covered by the Go unit tests (which stub the
# syscall) rather than here.

TestHaltHelp() {
    halt --help
}

TestPoweroffHelp() {
    poweroff --help
}

TestRebootHelp() {
    reboot --help
}

TestHaltVersion() {
    halt --version
}
