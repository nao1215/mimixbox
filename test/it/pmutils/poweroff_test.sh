# Dedicated integration helper for the 'poweroff' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
PoweroffHelp() {
    env -- 'poweroff' --help
}
