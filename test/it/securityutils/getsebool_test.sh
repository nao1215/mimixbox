# Dedicated integration helper for the 'getsebool' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
GetseboolHelp() {
    env -- 'getsebool' --help
}
