# Dedicated integration helper for the 'zip-pwcrack' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
ZipPwcrackHelp() {
    env -- 'zip-pwcrack' --help
}
