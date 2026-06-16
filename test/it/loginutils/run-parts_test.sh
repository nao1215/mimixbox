# Dedicated integration helper for the 'run-parts' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
RunPartsHelp() {
    env -- 'run-parts' --help
}
