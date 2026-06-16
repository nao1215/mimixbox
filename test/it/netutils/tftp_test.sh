# Dedicated integration helper for the 'tftp' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
TftpHelp() {
    env -- 'tftp' --help
}
