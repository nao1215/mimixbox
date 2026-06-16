# Dedicated integration helper for the 'tunctl' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
TunctlHelp() {
    env -- 'tunctl' --help
}
