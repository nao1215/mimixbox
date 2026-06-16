# Dedicated integration helper for the 'mkfs.vfat' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
MkfsVfatHelp() {
    env -- 'mkfs.vfat' --help
}
