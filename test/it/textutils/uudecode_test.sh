# Dedicated integration helper for the 'uudecode' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
UudecodeHelp() {
    env -- 'uudecode' --help
}
