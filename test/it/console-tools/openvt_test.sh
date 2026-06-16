# Dedicated integration helper for the 'openvt' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
OpenvtHelp() {
    env -- 'openvt' --help
}
