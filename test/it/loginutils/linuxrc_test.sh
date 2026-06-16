# Dedicated integration helper for the 'linuxrc' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
LinuxrcHelp() {
    env -- 'linuxrc' --help
}
