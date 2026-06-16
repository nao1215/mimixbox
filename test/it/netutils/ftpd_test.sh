# Dedicated integration helper for the 'ftpd' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
FtpdHelp() {
    env -- 'ftpd' --help
}
