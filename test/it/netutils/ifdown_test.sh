# Dedicated integration helper for the 'ifdown' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
IfdownHelp() {
    env -- 'ifdown' --help
}
