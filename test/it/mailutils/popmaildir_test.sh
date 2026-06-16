# Dedicated integration helper for the 'popmaildir' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
PopmaildirHelp() {
    env -- 'popmaildir' --help
}
