# Dedicated integration helper for the 'udhcpd' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
UdhcpdHelp() {
    env -- 'udhcpd' --help
}
