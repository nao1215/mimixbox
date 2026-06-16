# Dedicated integration helper for the 'nbd-client' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
NbdClientHelp() {
    env -- 'nbd-client' --help
}
