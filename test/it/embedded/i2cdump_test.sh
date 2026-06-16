# Dedicated integration helper for the 'i2cdump' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
I2cdumpHelp() {
    env -- 'i2cdump' --help
}
