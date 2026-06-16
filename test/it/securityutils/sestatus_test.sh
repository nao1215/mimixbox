# Dedicated integration helper for the 'sestatus' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
SestatusHelp() {
    env -- 'sestatus' --help
}
