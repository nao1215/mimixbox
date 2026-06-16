# Dedicated integration helper for the 'udhcpc' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
UdhcpcHelp() {
    env -- 'udhcpc' --help
}
