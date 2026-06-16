# Dedicated integration helper for the 'seedrng' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
SeedrngHelp() {
    env -- 'seedrng' --help
}
