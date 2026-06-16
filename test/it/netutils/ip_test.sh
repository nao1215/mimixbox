# Dedicated integration helper for the 'ip' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
IpHelp() {
    env -- 'ip' --help
}
