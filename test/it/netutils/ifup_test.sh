# Dedicated integration helper for the 'ifup' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
IfupHelp() {
    env -- 'ifup' --help
}
