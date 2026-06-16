# Dedicated integration helper for the 'setsid' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
SetsidHelp() {
    env -- 'setsid' --help
}
