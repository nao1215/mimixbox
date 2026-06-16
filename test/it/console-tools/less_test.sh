# Dedicated integration helper for the 'less' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
LessHelp() {
    env -- 'less' --help
}
