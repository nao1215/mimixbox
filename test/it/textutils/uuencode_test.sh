# Dedicated integration helper for the 'uuencode' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
UuencodeHelp() {
    env -- 'uuencode' --help
}
