# Dedicated integration helper for the 'pipe_progress' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
PipeProgressHelp() {
    env -- 'pipe_progress' --help
}
