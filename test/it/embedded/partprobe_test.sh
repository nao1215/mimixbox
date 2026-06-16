# Dedicated integration helper for the 'partprobe' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
PartprobeHelp() {
    env -- 'partprobe' --help
}
