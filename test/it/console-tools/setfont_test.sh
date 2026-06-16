# Dedicated integration helper for the 'setfont' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
SetfontHelp() {
    env -- 'setfont' --help
}
