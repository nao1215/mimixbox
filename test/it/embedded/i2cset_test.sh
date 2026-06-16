# Dedicated integration helper for the 'i2cset' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
I2csetHelp() {
    env -- 'i2cset' --help
}
