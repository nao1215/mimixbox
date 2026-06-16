# Dedicated integration helper for the 'lpq' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
LpqHelp() {
    env -- 'lpq' --help
}
