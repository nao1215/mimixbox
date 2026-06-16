# Dedicated integration helper for the 'xz' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
XzHelp() {
    env -- 'xz' --help
}
