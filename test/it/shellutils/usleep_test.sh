# Dedicated integration helper for the 'usleep' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
UsleepHelp() {
    env -- 'usleep' --help
}
