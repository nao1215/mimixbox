# Dedicated integration helper for the 'lzma' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
LzmaHelp() {
    env -- 'lzma' --help
}
