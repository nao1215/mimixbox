# Dedicated integration helper for the 'pscan' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
PscanHelp() {
    env -- 'pscan' --help
}
