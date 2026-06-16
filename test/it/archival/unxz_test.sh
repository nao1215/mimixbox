# Dedicated integration helper for the 'unxz' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
UnxzHelp() {
    env -- 'unxz' --help
}
