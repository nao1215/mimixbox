# Dedicated integration helper for the 'unit' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
UnitHelp() {
    env -- 'unit' --help
}
