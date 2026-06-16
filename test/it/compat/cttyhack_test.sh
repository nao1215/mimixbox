# Dedicated integration helper for the 'cttyhack' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
CttyhackHelp() {
    env -- 'cttyhack' --help
}
