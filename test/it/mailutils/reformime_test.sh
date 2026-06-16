# Dedicated integration helper for the 'reformime' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
ReformimeHelp() {
    env -- 'reformime' --help
}
