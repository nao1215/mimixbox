# Dedicated integration helper for the 'ether-wake' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
EtherWakeHelp() {
    env -- 'ether-wake' --help
}
