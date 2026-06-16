# Dedicated integration helper for the 'fgrep' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
FgrepHelp() {
    env -- 'fgrep' --help
}
