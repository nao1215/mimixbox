# Dedicated integration helper for the 'ifplugd' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
IfplugdHelp() {
    env -- 'ifplugd' --help
}
