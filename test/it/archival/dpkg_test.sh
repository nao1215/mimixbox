# Dedicated integration helper for the 'dpkg' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
DpkgHelp() {
    env -- 'dpkg' --help
}
