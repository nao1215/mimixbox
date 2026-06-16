# Dedicated integration helper for the 'ftpget' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
FtpgetHelp() {
    env -- 'ftpget' --help
}
