# Dedicated integration helper for the 'i2cdetect' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
I2cdetectHelp() {
    env -- 'i2cdetect' --help
}
