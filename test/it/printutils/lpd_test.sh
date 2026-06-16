# Dedicated integration helper for the 'lpd' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
LpdHelp() {
    env -- 'lpd' --help
}
