# Dedicated integration helper for the 'arp' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
ArpHelp() {
    env -- 'arp' --help
}
