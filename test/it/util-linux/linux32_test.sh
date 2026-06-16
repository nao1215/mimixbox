# Dedicated integration helper for the 'linux32' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
Linux32Help() {
    env -- 'linux32' --help
}
