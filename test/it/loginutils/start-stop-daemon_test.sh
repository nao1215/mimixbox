# Dedicated integration helper for the 'start-stop-daemon' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
StartStopDaemonHelp() {
    env -- 'start-stop-daemon' --help
}
