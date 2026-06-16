# Dedicated integration helper for the 'http-status-code' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
HttpStatusCodeHelp() {
    env -- 'http-status-code' --help
}
