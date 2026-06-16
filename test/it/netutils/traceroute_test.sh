# Dedicated integration helper for the 'traceroute' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
TracerouteHelp() {
    env -- 'traceroute' --help
}
