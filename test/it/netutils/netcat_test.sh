# Dedicated integration helper for the 'netcat' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
NetcatHelp() {
    env -- 'netcat' --help
}
