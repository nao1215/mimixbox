# Dedicated integration helper for the 'reboot' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
RebootHelp() {
    env -- 'reboot' --help
}
