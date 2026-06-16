# Dedicated integration helper for the 'sum' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
SumHelp() {
    env -- 'sum' --help
}
