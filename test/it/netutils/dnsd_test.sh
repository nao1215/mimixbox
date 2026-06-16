# Dedicated integration helper for the 'dnsd' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
DnsdHelp() {
    env -- 'dnsd' --help
}
