# Dedicated integration helper for the 'lzopcat' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
LzopcatHelp() {
    env -- 'lzopcat' --help
}
