# Dedicated integration helper for the 'restorecon' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
RestoreconHelp() {
    env -- 'restorecon' --help
}
