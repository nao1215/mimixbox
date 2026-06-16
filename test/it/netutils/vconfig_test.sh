# Dedicated integration helper for the 'vconfig' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
VconfigHelp() {
    env -- 'vconfig' --help
}
