# Dedicated integration helper for the 'loadfont' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
LoadfontHelp() {
    env -- 'loadfont' --help
}
