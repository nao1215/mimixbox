# Dedicated integration helper for the 'zcat' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
ZcatHelp() {
    env -- 'zcat' --help
}
