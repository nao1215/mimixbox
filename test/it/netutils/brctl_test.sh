# Dedicated integration helper for the 'brctl' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
BrctlHelp() {
    env -- 'brctl' --help
}
