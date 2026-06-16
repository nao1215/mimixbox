# Dedicated integration helper for the 'log-collect' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
LogCollectHelp() {
    env -- 'log-collect' --help
}
