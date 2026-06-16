# Dedicated integration helper for the 'readahead' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
ReadaheadHelp() {
    env -- 'readahead' --help
}
