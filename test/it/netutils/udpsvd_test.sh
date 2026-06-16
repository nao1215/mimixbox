# Dedicated integration helper for the 'udpsvd' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
UdpsvdHelp() {
    env -- 'udpsvd' --help
}
