# Dedicated integration helper for the 'sha3sum' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
Sha3sumHelp() {
    env -- 'sha3sum' --help
}
