# Dedicated integration helper for the 'sddf' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
SddfHelp() {
    env -- 'sddf' --help
}
