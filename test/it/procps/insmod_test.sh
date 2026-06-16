# Dedicated integration helper for the 'insmod' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
InsmodHelp() {
    env -- 'insmod' --help
}
