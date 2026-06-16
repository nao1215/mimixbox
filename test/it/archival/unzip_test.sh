# Dedicated integration helper for the 'unzip' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
UnzipHelp() {
    env -- 'unzip' --help
}
