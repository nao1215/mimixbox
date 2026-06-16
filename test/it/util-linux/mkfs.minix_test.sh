# Dedicated integration helper for the 'mkfs.minix' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
MkfsMinixHelp() {
    env -- 'mkfs.minix' --help
}
