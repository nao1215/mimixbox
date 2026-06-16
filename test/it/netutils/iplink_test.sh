# Dedicated integration helper for the 'iplink' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
IplinkHelp() {
    env -- 'iplink' --help
}
