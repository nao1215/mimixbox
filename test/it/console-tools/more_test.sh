# Dedicated integration helper for the 'more' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
MoreHelp() {
    env -- 'more' --help
}
