# Dedicated integration helper for the 'mkfs.ext2' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
MkfsExt2Help() {
    env -- 'mkfs.ext2' --help
}
