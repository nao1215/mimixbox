# Dedicated integration helper for the 'sha384sum' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
Sha384sumHelp() {
    env -- 'sha384sum' --help
}
