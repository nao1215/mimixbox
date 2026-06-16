# Dedicated integration helper for the 'udhcpc6' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
Udhcpc6Help() {
    env -- 'udhcpc6' --help
}
