# Dedicated integration helper for the 'linux64' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
Linux64Help() {
    env -- 'linux64' --help
}
