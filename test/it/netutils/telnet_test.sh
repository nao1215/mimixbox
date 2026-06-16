# Dedicated integration helper for the 'telnet' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
TelnetHelp() {
    env -- 'telnet' --help
}
