# Dedicated integration helper for the 'run-init' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
RunInitHelp() {
    env -- 'run-init' --help
}
