# Dedicated integration helper for the 'xzcat' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
XzcatHelp() {
    env -- 'xzcat' --help
}
