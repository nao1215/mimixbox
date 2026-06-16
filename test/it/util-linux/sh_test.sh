# Dedicated integration helper for the 'sh' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
ShHelp() {
    env -- 'sh' --help
}
