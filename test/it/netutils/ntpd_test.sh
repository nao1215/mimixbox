# Dedicated integration helper for the 'ntpd' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
NtpdHelp() {
    env -- 'ntpd' --help
}
