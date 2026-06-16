# Dedicated integration helper for the 'traceroute6' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
Traceroute6Help() {
    env -- 'traceroute6' --help
}
