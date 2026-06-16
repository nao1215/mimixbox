# Dedicated integration helper for the 'iprule' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
IpruleHelp() {
    env -- 'iprule' --help
}
