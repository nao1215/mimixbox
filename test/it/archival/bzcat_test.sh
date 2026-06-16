# Dedicated integration helper for the 'bzcat' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
BzcatHelp() {
    env -- 'bzcat' --help
}
