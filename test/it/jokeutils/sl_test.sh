# Dedicated integration helper for the 'sl' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
SlHelp() {
    env -- 'sl' --help
}
