# Dedicated integration helper for the 'load_policy' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
LoadPolicyHelp() {
    env -- 'load_policy' --help
}
