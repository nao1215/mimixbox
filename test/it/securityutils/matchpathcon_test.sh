# Dedicated integration helper for the 'matchpathcon' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
MatchpathconHelp() {
    env -- 'matchpathcon' --help
}
