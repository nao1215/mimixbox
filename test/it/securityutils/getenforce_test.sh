# Dedicated integration helper for the 'getenforce' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
GetenforceHelp() {
    env -- 'getenforce' --help
}
