# Dedicated integration helper for the 'chcon' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
ChconHelp() {
    env -- 'chcon' --help
}
