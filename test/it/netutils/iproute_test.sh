# Dedicated integration helper for the 'iproute' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
IprouteHelp() {
    env -- 'iproute' --help
}
