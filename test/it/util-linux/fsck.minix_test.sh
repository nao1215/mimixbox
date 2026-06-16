# Dedicated integration helper for the 'fsck.minix' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
FsckMinixHelp() {
    env -- 'fsck.minix' --help
}
