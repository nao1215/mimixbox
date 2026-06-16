# Dedicated integration helper for the 'dumpkmap' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
DumpkmapHelp() {
    env -- 'dumpkmap' --help
}
