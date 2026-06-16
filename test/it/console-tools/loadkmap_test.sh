# Dedicated integration helper for the 'loadkmap' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
LoadkmapHelp() {
    env -- 'loadkmap' --help
}
