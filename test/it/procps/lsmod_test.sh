# Dedicated integration helper for the 'lsmod' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
LsmodHelp() {
    env -- 'lsmod' --help
}
