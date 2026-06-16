# Dedicated integration helper for the 'watchdog' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
WatchdogHelp() {
    env -- 'watchdog' --help
}
