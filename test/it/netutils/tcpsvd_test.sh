# Dedicated integration helper for the 'tcpsvd' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
TcpsvdHelp() {
    env -- 'tcpsvd' --help
}
