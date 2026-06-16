# Dedicated integration helper for the 'busybox' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
BusyboxHelp() {
    env -- 'busybox' --help
}
