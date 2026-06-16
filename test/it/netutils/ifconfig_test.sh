# Dedicated integration helper for the 'ifconfig' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
IfconfigHelp() {
    env -- 'ifconfig' --help
}
