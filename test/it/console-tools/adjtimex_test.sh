# Dedicated integration helper for the 'adjtimex' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
AdjtimexHelp() {
    env -- 'adjtimex' --help
}
