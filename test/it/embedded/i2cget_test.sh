# Dedicated integration helper for the 'i2cget' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
I2cgetHelp() {
    env -- 'i2cget' --help
}
