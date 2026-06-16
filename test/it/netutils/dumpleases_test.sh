# Dedicated integration helper for the 'dumpleases' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
DumpleasesHelp() {
    env -- 'dumpleases' --help
}
