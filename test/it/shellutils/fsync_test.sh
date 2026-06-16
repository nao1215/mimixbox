# Dedicated integration helper for the 'fsync' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
FsyncHelp() {
    env -- 'fsync' --help
}
