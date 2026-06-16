# Dedicated integration helper for the 'depmod' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
DepmodHelp() {
    env -- 'depmod' --help
}
