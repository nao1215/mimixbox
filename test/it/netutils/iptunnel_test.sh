# Dedicated integration helper for the 'iptunnel' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
IptunnelHelp() {
    env -- 'iptunnel' --help
}
