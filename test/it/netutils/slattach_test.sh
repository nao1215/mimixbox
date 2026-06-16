# Dedicated integration helper for the 'slattach' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
SlattachHelp() {
    env -- 'slattach' --help
}
