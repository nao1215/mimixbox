# Dedicated integration helper for the 'rx' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
RxHelp() {
    env -- 'rx' --help
}
