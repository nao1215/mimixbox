# Dedicated integration helper for the '[' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
BracketHelp() {
    env -- '[' --help
}
