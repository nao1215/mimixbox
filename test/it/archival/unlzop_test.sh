# Dedicated integration helper for the 'unlzop' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
UnlzopHelp() {
    env -- 'unlzop' --help
}
