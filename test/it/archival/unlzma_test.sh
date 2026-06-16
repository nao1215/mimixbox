# Dedicated integration helper for the 'unlzma' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
UnlzmaHelp() {
    env -- 'unlzma' --help
}
