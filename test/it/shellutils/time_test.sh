# Dedicated integration helper for the 'time' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
TimeHelp() {
    env -- 'time' --help
}
