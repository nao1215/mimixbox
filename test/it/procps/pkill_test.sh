# Dedicated integration helper for the 'pkill' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
PkillHelp() {
    env -- 'pkill' --help
}
