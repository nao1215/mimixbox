# Dedicated integration helper for the 'egrep' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
EgrepHelp() {
    env -- 'egrep' --help
}
