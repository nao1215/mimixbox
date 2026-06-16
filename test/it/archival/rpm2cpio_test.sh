# Dedicated integration helper for the 'rpm2cpio' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
Rpm2cpioHelp() {
    env -- 'rpm2cpio' --help
}
