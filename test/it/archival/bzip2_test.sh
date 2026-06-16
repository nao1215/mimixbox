# Dedicated integration helper for the 'bzip2' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
Bzip2Help() {
    env -- 'bzip2' --help
}
