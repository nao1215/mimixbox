# Dedicated integration helper for the 'chown' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
ChownHelp() {
    env -- 'chown' --help
}
