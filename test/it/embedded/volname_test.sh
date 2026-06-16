# Dedicated integration helper for the 'volname' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
VolnameHelp() {
    env -- 'volname' --help
}
