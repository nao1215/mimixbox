# Dedicated integration helper for the 'selinuxenabled' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
SelinuxenabledHelp() {
    env -- 'selinuxenabled' --help
}
