# Dedicated integration helper for the 'bash' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
BashHelp() {
    env -- 'bash' --help
}
