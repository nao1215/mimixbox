# Dedicated integration helper for the 'ftpput' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
FtpputHelp() {
    env -- 'ftpput' --help
}
