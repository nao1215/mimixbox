# Dedicated integration helper for the 'fallocate' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
FallocateHelp() {
    env -- 'fallocate' --help
}
