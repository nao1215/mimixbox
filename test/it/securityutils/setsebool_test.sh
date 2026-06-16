# Dedicated integration helper for the 'setsebool' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
SetseboolHelp() {
    env -- 'setsebool' --help
}
