# Dedicated integration helper for the 'setfiles' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
SetfilesHelp() {
    env -- 'setfiles' --help
}
