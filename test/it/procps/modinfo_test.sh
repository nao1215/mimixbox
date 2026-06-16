# Dedicated integration helper for the 'modinfo' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
ModinfoHelp() {
    env -- 'modinfo' --help
}
