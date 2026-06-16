# Dedicated integration helper for the 'rmmod' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
RmmodHelp() {
    env -- 'rmmod' --help
}
