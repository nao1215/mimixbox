# Dedicated integration helper for the 'inetd' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
InetdHelp() {
    env -- 'inetd' --help
}
