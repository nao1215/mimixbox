# Dedicated integration helper for the 'zcip' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
ZcipHelp() {
    env -- 'zcip' --help
}
