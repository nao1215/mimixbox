# Dedicated integration helper for the 'setenforce' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
SetenforceHelp() {
    env -- 'setenforce' --help
}
