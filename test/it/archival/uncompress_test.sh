# Dedicated integration helper for the 'uncompress' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
UncompressHelp() {
    env -- 'uncompress' --help
}
