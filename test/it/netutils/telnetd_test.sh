# Dedicated integration helper for the 'telnetd' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
TelnetdHelp() {
    env -- 'telnetd' --help
}
