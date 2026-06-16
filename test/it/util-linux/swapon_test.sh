# Dedicated integration helper for the 'swapon' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
SwaponHelp() {
    env -- 'swapon' --help
}
