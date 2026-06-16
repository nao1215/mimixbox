# Dedicated integration helper for the 'raidautorun' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
RaidautorunHelp() {
    env -- 'raidautorun' --help
}
