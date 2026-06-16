# Dedicated integration helper for the 'uptime' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
UptimeHelp() {
    env -- 'uptime' --help
}
