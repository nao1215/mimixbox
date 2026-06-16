# Dedicated integration helper for the 'devmem' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
DevmemHelp() {
    env -- 'devmem' --help
}
