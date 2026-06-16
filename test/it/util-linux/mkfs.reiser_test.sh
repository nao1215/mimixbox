# Dedicated integration helper for the 'mkfs.reiser' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
MkfsReiserHelp() {
    env -- 'mkfs.reiser' --help
}
