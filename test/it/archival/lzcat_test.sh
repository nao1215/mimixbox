# Dedicated integration helper for the 'lzcat' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
LzcatHelp() {
    env -- 'lzcat' --help
}
