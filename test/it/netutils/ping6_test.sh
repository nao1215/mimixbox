# Dedicated integration helper for the 'ping6' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
Ping6Help() {
    env -- 'ping6' --help
}
