# Dedicated integration helper for the 'ash' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
AshHelp() {
    env -- 'ash' --help
}
