# Dedicated integration helper for the 'tc' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
TcHelp() {
    env -- 'tc' --help
}
