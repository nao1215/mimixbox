# Dedicated integration helper for the 'netstat' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
NetstatHelp() {
    env -- 'netstat' --help
}
