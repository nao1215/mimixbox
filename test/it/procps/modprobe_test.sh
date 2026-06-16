# Dedicated integration helper for the 'modprobe' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
ModprobeHelp() {
    env -- 'modprobe' --help
}
