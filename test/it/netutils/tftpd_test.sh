# Dedicated integration helper for the 'tftpd' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
TftpdHelp() {
    env -- 'tftpd' --help
}
