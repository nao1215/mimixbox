# Dedicated integration helper for the 'mkdosfs' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
MkdosfsHelp() {
    env -- 'mkdosfs' --help
}
