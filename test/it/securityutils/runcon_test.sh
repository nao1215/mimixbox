# Dedicated integration helper for the 'runcon' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
RunconHelp() {
    env -- 'runcon' --help
}
