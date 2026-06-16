# Dedicated integration helper for the 'pwdx' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
PwdxHelp() {
    env -- 'pwdx' --help
}
