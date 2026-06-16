# Dedicated integration helper for the 'conspy' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
ConspyHelp() {
    env -- 'conspy' --help
}
