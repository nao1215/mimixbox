# Dedicated integration helper for the 'addgroup' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
AddgroupHelp() {
    env -- 'addgroup' --help
}
