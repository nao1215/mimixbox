# Dedicated integration helper for the 'arping' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
ArpingHelp() {
    env -- 'arping' --help
}
