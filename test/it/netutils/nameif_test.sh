# Dedicated integration helper for the 'nameif' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
NameifHelp() {
    env -- 'nameif' --help
}
