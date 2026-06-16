# Dedicated integration helper for the 'dhcprelay' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
DhcprelayHelp() {
    env -- 'dhcprelay' --help
}
