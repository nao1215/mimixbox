# Dedicated integration helper for the 'ipaddr' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
IpaddrHelp() {
    env -- 'ipaddr' --help
}
