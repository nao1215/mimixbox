# Dedicated integration helper for the 'lpr' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
LprHelp() {
    env -- 'lpr' --help
}
