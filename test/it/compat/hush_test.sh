# Dedicated integration helper for the 'hush' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
HushHelp() {
    env -- 'hush' --help
}
